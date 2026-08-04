package blacklist

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
	"github.com/redis/go-redis/v9"
)

type EntryStore interface {
	ActiveEntries(ctx context.Context) ([]Entry, error)
}

type Service struct {
	store EntryStore
	redis *redis.Client
	now   func() time.Time
}

func NewService(store EntryStore, redisClient *redis.Client) *Service {
	return &Service{store: store, redis: redisClient, now: time.Now}
}

func (s *Service) WarmCache(ctx context.Context) error {
	entries, err := s.store.ActiveEntries(ctx)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		_ = s.redis.Set(ctx, cacheKey(entry.Type, entry.Value), "1", 0).Err()
	}
	return nil
}

func (s *Service) Check(ctx context.Context, event contracts.TransactionEnrichedEvent) contracts.BlacklistCheckEvent {
	matches := make([]string, 0, 4)
	if event.DeviceID != "" && s.blacklisted(ctx, "DEVICE_ID", event.DeviceID) {
		matches = append(matches, "DEVICE_ID:"+event.DeviceID)
	}
	if event.SenderUserID != 0 && s.blacklisted(ctx, "USER_ID", fmt.Sprintf("%d", event.SenderUserID)) {
		matches = append(matches, fmt.Sprintf("SENDER_USER_ID:%d", event.SenderUserID))
	}
	if event.ReceiverUserID != nil && s.blacklisted(ctx, "USER_ID", fmt.Sprintf("%d", *event.ReceiverUserID)) {
		matches = append(matches, fmt.Sprintf("RECEIVER_USER_ID:%d", *event.ReceiverUserID))
	}
	if event.MerchantID != nil && s.blacklisted(ctx, "MERCHANT_ID", fmt.Sprintf("%d", *event.MerchantID)) {
		matches = append(matches, fmt.Sprintf("MERCHANT_ID:%d", *event.MerchantID))
	}

	var reason *string
	var hint *string
	if len(matches) > 0 {
		value := reasonFor(matches)
		reason = &value
		block := "BLOCK"
		hint = &block
	}
	return contracts.BlacklistCheckEvent{
		TransactionID: event.TransactionID,
		BlacklistHit:  len(matches) > 0,
		Reason:        reason,
		DecisionHint:  hint,
		Transaction:   event,
		EvaluatedAt:   s.now().UTC(),
	}
}

func (s *Service) blacklisted(ctx context.Context, typ string, value string) bool {
	raw, err := s.redis.Get(ctx, cacheKey(typ, value)).Result()
	return err == nil && raw == "1"
}

func cacheKey(typ string, value string) string {
	switch strings.ToUpper(typ) {
	case "DEVICE_ID":
		return "blacklist:device:" + value
	case "USER_ID":
		return "blacklist:user:" + value
	case "MERCHANT_ID":
		return "blacklist:merchant:" + value
	default:
		return "blacklist:account:" + value
	}
}

func reasonFor(matches []string) string {
	for _, match := range matches {
		switch {
		case strings.HasPrefix(match, "DEVICE_ID:"):
			return "BLACKLISTED_DEVICE"
		case strings.HasPrefix(match, "MERCHANT_ID:"):
			return "BLACKLISTED_MERCHANT"
		case strings.HasPrefix(match, "SENDER_USER_ID:"):
			return "BLACKLISTED_SENDER"
		case strings.HasPrefix(match, "RECEIVER_USER_ID:"):
			return "BLACKLISTED_RECEIVER"
		}
	}
	return "BLACKLISTED_RECEIVER"
}
