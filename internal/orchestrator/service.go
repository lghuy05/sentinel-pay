package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

const aggregationTTL = 15 * time.Second

type Service struct {
	redis  *redis.Client
	store  *Store
	writer DecisionPublisher
	now    func() time.Time
}

type DecisionPublisher interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}

func NewService(redisClient *redis.Client, store *Store, writer DecisionPublisher) *Service {
	return &Service{redis: redisClient, store: store, writer: writer, now: time.Now}
}

func (s *Service) HandleBlacklist(ctx context.Context, event contracts.BlacklistCheckEvent) error {
	if event.TransactionID == "" {
		return nil
	}
	features := featureMap(event.Transaction)
	if event.BlacklistHit {
		return s.finalize(ctx, event.TransactionID, finalDecisionInput{
			finalDecision:  contracts.FraudDecisionBlock,
			decisionReason: "BLACKLIST",
			blacklistHit:   boolPtr(true),
			features:       features,
		})
	}
	if len(features) > 0 {
		if err := s.storeTransactionSnapshot(ctx, event.TransactionID, features); err != nil {
			return err
		}
	}
	return s.upsertAggregation(ctx, event.TransactionID, map[string]string{"blacklistHit": "false"})
}

func (s *Service) HandleRule(ctx context.Context, event contracts.RuleEvaluationEvent) error {
	if event.TransactionID == "" {
		return nil
	}
	switch event.RuleBand {
	case "SAFE":
		return s.finalize(ctx, event.TransactionID, finalDecisionInputFromRule(event, "RULE_SAFE", contracts.FraudDecisionAllow))
	case "RISK":
		return s.finalize(ctx, event.TransactionID, finalDecisionInputFromRule(event, "RULE_RISK", contracts.FraudDecisionBlock))
	}
	if len(event.Features) > 0 {
		if err := s.storeTransactionSnapshot(ctx, event.TransactionID, event.Features); err != nil {
			return err
		}
	}
	matches, _ := json.Marshal(emptyListIfNil(event.RuleMatches))
	updates := map[string]string{
		"ruleScore":   strconv.FormatFloat(event.RuleScore, 'f', -1, 64),
		"ruleBand":    event.RuleBand,
		"ruleMatches": string(matches),
	}
	if event.RuleVersion != nil {
		updates["ruleVersion"] = strconv.Itoa(*event.RuleVersion)
	}
	if err := s.upsertAggregation(ctx, event.TransactionID, updates); err != nil {
		return err
	}
	if event.RuleBand != "GRAY" {
		return nil
	}
	mlScore, err := s.redis.HGet(ctx, aggregateKey(event.TransactionID), "mlScore").Result()
	if err != nil || mlScore == "" {
		return nil
	}
	score := parseFloatPtr(mlScore)
	if score == nil {
		return nil
	}
	mlBand := "GRAY"
	decision := contracts.FraudDecisionHold
	reason := "ML_GRAY"
	if *score < 0.30 {
		mlBand = "SAFE"
		decision = contracts.FraudDecisionAllow
		reason = "ML_SAFE"
	} else if *score > 0.70 {
		mlBand = "RISK"
		decision = contracts.FraudDecisionBlock
		reason = "ML_RISK"
	}
	modelVersion, _ := s.redis.HGet(ctx, aggregateKey(event.TransactionID), "modelVersion").Result()
	return s.finalize(ctx, event.TransactionID, finalDecisionInput{
		finalDecision:  decision,
		decisionReason: reason,
		blacklistHit:   boolPtr(false),
		ruleScore:      &event.RuleScore,
		ruleBand:       &event.RuleBand,
		ruleMatches:    event.RuleMatches,
		mlScore:        score,
		mlBand:         &mlBand,
		modelVersion:   blankToNil(modelVersion),
		ruleVersion:    event.RuleVersion,
		features:       event.Features,
	})
}

func (s *Service) HandleML(ctx context.Context, event contracts.MLScoreEvent) error {
	if event.TransactionID == "" {
		return nil
	}
	updates := map[string]string{
		"mlScore": strconv.FormatFloat(event.MLScore, 'f', -1, 64),
	}
	if event.ModelVersion != "" {
		updates["modelVersion"] = event.ModelVersion
	}
	if err := s.upsertAggregation(ctx, event.TransactionID, updates); err != nil {
		return err
	}
	band, err := s.redis.HGet(ctx, aggregateKey(event.TransactionID), "ruleBand").Result()
	if err != nil || band != "GRAY" {
		return nil
	}
	mlBand := "GRAY"
	decision := contracts.FraudDecisionHold
	reason := "ML_GRAY"
	if event.MLScore < 0.30 {
		mlBand = "SAFE"
		decision = contracts.FraudDecisionAllow
		reason = "ML_SAFE"
	} else if event.MLScore > 0.70 {
		mlBand = "RISK"
		decision = contracts.FraudDecisionBlock
		reason = "ML_RISK"
	}
	return s.finalize(ctx, event.TransactionID, finalDecisionInput{
		finalDecision:  decision,
		decisionReason: reason,
		blacklistHit:   boolPtr(false),
		ruleBand:       stringPtr("GRAY"),
		mlScore:        &event.MLScore,
		mlBand:         &mlBand,
		modelVersion:   blankToNil(event.ModelVersion),
	})
}

func (s *Service) finalize(ctx context.Context, transactionID string, input finalDecisionInput) error {
	first, err := s.redis.SetNX(ctx, "fraud:finalized:"+transactionID, "1", 2*time.Minute).Result()
	if err != nil {
		return err
	}
	if !first {
		return nil
	}

	values := map[string]string{}
	if input.needsAggregation() {
		values, _ = s.redis.HGetAll(ctx, aggregateKey(transactionID)).Result()
	}
	features := input.features
	if len(features) == 0 {
		features = deserializeMap(values["featuresJson"])
	}
	senderUserID := parseInt64Feature(values["senderUserId"], features, "senderUserId")
	receiverUserID := parseInt64Feature(values["receiverUserId"], features, "receiverUserId")
	merchantID := parseInt64Feature(values["merchantId"], features, "merchantId")
	amount := parseFloatFeature(values["amount"], features, "amount")
	currency := parseStringFeature(values["currency"], features, "currency")
	country := parseStringFeature(values["country"], features, "senderAccountCountry")
	accountID := senderUserID

	blacklistHit := false
	if input.blacklistHit != nil {
		blacklistHit = *input.blacklistHit
	} else {
		blacklistHit = values["blacklistHit"] == "true"
	}
	ruleScore := input.ruleScore
	if ruleScore == nil {
		ruleScore = parseFloatPtr(values["ruleScore"])
	}
	ruleBand := input.ruleBand
	if ruleBand == nil {
		ruleBand = blankToNil(values["ruleBand"])
	}
	ruleMatches := input.ruleMatches
	if ruleMatches == nil {
		ruleMatches = deserializeList(values["ruleMatches"])
	}
	mlScore := input.mlScore
	if mlScore == nil {
		mlScore = parseFloatPtr(values["mlScore"])
	}
	modelVersion := input.modelVersion
	if modelVersion == nil {
		modelVersion = blankToNil(values["modelVersion"])
	}
	ruleVersion := input.ruleVersion
	if ruleVersion == nil {
		ruleVersion = parseIntPtr(values["ruleVersion"])
	}
	featuresJSON := serializeMap(features)
	out := contracts.FraudFinalDecisionEvent{
		TransactionID:  transactionID,
		AccountID:      accountID,
		SenderUserID:   senderUserID,
		ReceiverUserID: receiverUserID,
		MerchantID:     merchantID,
		Amount:         amount,
		Currency:       currency,
		Country:        country,
		FeaturesJSON:   featuresJSON,
		BlacklistHit:   blacklistHit,
		RuleScore:      ruleScore,
		RuleBand:       ruleBand,
		RuleMatches:    emptyListIfNil(ruleMatches),
		MLScore:        mlScore,
		MLBand:         input.mlBand,
		FinalDecision:  input.finalDecision,
		DecisionReason: input.decisionReason,
		ModelVersion:   modelVersion,
		RuleVersion:    ruleVersion,
		DecidedAt:      s.now().UTC(),
	}
	if err := s.store.Save(ctx, out); err != nil {
		return err
	}
	payload, err := json.Marshal(out)
	if err != nil {
		return err
	}
	if err := s.writer.WriteMessages(ctx, kafka.Message{Key: []byte(transactionID), Value: payload}); err != nil {
		return err
	}
	_ = s.redis.Del(ctx, aggregateKey(transactionID)).Err()
	return nil
}

func (s *Service) storeTransactionSnapshot(ctx context.Context, transactionID string, features map[string]any) error {
	return s.upsertAggregation(ctx, transactionID, map[string]string{"featuresJson": serializeMap(features)})
}

func (s *Service) upsertAggregation(ctx context.Context, transactionID string, updates map[string]string) error {
	if len(updates) == 0 {
		return nil
	}
	key := aggregateKey(transactionID)
	if err := s.redis.HSet(ctx, key, updates).Err(); err != nil {
		return err
	}
	return s.redis.Expire(ctx, key, aggregationTTL).Err()
}

type finalDecisionInput struct {
	finalDecision  contracts.FraudDecision
	decisionReason string
	blacklistHit   *bool
	ruleScore      *float64
	ruleBand       *string
	ruleMatches    []string
	mlScore        *float64
	mlBand         *string
	modelVersion   *string
	ruleVersion    *int
	features       map[string]any
}

func (i finalDecisionInput) needsAggregation() bool {
	return len(i.features) == 0 || i.blacklistHit == nil || i.ruleScore == nil || i.ruleBand == nil ||
		i.ruleMatches == nil || i.mlScore == nil || i.modelVersion == nil || i.ruleVersion == nil
}

func finalDecisionInputFromRule(event contracts.RuleEvaluationEvent, reason string, decision contracts.FraudDecision) finalDecisionInput {
	return finalDecisionInput{
		finalDecision:  decision,
		decisionReason: reason,
		blacklistHit:   boolPtr(false),
		ruleScore:      &event.RuleScore,
		ruleBand:       &event.RuleBand,
		ruleMatches:    event.RuleMatches,
		ruleVersion:    event.RuleVersion,
		features:       event.Features,
	}
}

func aggregateKey(transactionID string) string {
	return "fraud:aggregate:" + transactionID
}

func featureMap(event contracts.TransactionEnrichedEvent) map[string]any {
	payload, err := json.Marshal(event)
	if err != nil {
		return map[string]any{}
	}
	return deserializeMap(string(payload))
}

func serializeMap(values map[string]any) string {
	if values == nil {
		values = map[string]any{}
	}
	payload, err := json.Marshal(values)
	if err != nil {
		return "{}"
	}
	return string(payload)
}

func deserializeMap(raw string) map[string]any {
	if raw == "" {
		return map[string]any{}
	}
	var values map[string]any
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return map[string]any{}
	}
	return values
}

func deserializeList(raw string) []string {
	if raw == "" {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return []string{}
	}
	return values
}

func parseInt64Feature(raw string, features map[string]any, key string) *int64 {
	if value, ok := parseInt64(raw); ok {
		return &value
	}
	if features == nil {
		return nil
	}
	return int64FromAny(features[key])
}

func parseFloatFeature(raw string, features map[string]any, key string) *float64 {
	if value, ok := parseFloat(raw); ok {
		return &value
	}
	if features == nil {
		return nil
	}
	return floatFromAny(features[key])
}

func parseStringFeature(raw string, features map[string]any, key string) *string {
	if raw != "" {
		return &raw
	}
	if features == nil || features[key] == nil {
		return nil
	}
	value := fmt.Sprint(features[key])
	return &value
}

func int64FromAny(raw any) *int64 {
	switch value := raw.(type) {
	case float64:
		parsed := int64(value)
		return &parsed
	case int64:
		return &value
	case int:
		parsed := int64(value)
		return &parsed
	case string:
		if parsed, ok := parseInt64(value); ok {
			return &parsed
		}
	}
	return nil
}

func floatFromAny(raw any) *float64 {
	switch value := raw.(type) {
	case float64:
		return &value
	case int:
		parsed := float64(value)
		return &parsed
	case int64:
		parsed := float64(value)
		return &parsed
	case string:
		if parsed, ok := parseFloat(value); ok {
			return &parsed
		}
	}
	return nil
}

func parseFloatPtr(raw string) *float64 {
	if value, ok := parseFloat(raw); ok {
		return &value
	}
	return nil
}

func parseIntPtr(raw string) *int {
	if raw == "" {
		return nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &value
}

func parseFloat(raw string) (float64, bool) {
	if raw == "" {
		return 0, false
	}
	value, err := strconv.ParseFloat(raw, 64)
	return value, err == nil
}

func parseInt64(raw string) (int64, bool) {
	if raw == "" {
		return 0, false
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	return value, err == nil
}

func blankToNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func emptyListIfNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
