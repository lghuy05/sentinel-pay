package com.example.fraud_orchestrator.service;

import java.math.BigDecimal;
import java.time.Duration;
import java.time.Instant;
import java.util.Collections;
import java.util.List;
import java.util.Map;

import org.springframework.data.redis.core.HashOperations;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

import com.example.fraud_orchestrator.entity.FraudDecisionRecord;
import com.example.fraud_orchestrator.event.BlacklistCheckEvent;
import com.example.fraud_orchestrator.event.FraudDecision;
import com.example.fraud_orchestrator.event.FraudFinalDecisionEvent;
import com.example.fraud_orchestrator.event.MlScoreEvent;
import com.example.fraud_orchestrator.event.RuleEvaluationEvent;
import com.example.fraud_orchestrator.repository.FraudDecisionRepository;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;

@Service
public class FraudOrchestrationService {

    private static final String OUT_TOPIC = "fraud.final";
    private static final Duration AGGREGATION_TTL = Duration.ofSeconds(15);

    private static final String FIELD_BLACKLIST_HIT = "blacklistHit";
    private static final String FIELD_BLACKLIST_REASON = "blacklistReason";
    private static final String FIELD_RULE_SCORE = "ruleScore";
    private static final String FIELD_RULE_BAND = "ruleBand";
    private static final String FIELD_RULE_MATCHES = "ruleMatches";
    private static final String FIELD_RULE_VERSION = "ruleVersion";
    private static final String FIELD_ML_SCORE = "mlScore";
    private static final String FIELD_MODEL_VERSION = "modelVersion";
    private static final String FIELD_FEATURES = "featuresJson";
    private static final String FIELD_SENDER_USER_ID = "senderUserId";
    private static final String FIELD_RECEIVER_USER_ID = "receiverUserId";
    private static final String FIELD_MERCHANT_ID = "merchantId";
    private static final String FIELD_AMOUNT = "amount";
    private static final String FIELD_CURRENCY = "currency";
    private static final String FIELD_COUNTRY = "country";

    private final StringRedisTemplate redisTemplate;
    private final KafkaTemplate<String, FraudFinalDecisionEvent> kafkaTemplate;
    private final FraudDecisionRepository decisionRepository;
    private final ObjectMapper objectMapper;

    public FraudOrchestrationService(
            StringRedisTemplate redisTemplate,
            KafkaTemplate<String, FraudFinalDecisionEvent> kafkaTemplate,
            FraudDecisionRepository decisionRepository,
            ObjectMapper objectMapper
    ) {
        this.redisTemplate = redisTemplate;
        this.kafkaTemplate = kafkaTemplate;
        this.decisionRepository = decisionRepository;
        this.objectMapper = objectMapper;
    }

    public void handleBlacklistEvent(BlacklistCheckEvent event) {
        String transactionId = event.getTransactionId();
        if (transactionId == null) {
            return;
        }

        Map<String, Object> transaction = event.getTransaction();
        if (transaction != null) {
            storeTransactionSnapshot(transactionId, transaction);
        }

        if (event.isBlacklistHit()) {
            finalizeDecision(transactionId, FinalDecisionInput.fromBlacklist(event));
            return;
        }

        HashOperations<String, Object, Object> hashOps = redisTemplate.opsForHash();
        String key = aggregateKey(transactionId);
        hashOps.put(key, FIELD_BLACKLIST_HIT, "false");
        storeIfPresent(hashOps, key, FIELD_BLACKLIST_REASON, event.getReason());
        redisTemplate.expire(key, AGGREGATION_TTL);
    }

    public void handleRuleEvent(RuleEvaluationEvent event) {
        String transactionId = event.getTransactionId();
        if (transactionId == null) {
            return;
        }

        Map<String, Object> features = event.getFeatures();
        if (features != null) {
            storeTransactionSnapshot(transactionId, features);
        }

        HashOperations<String, Object, Object> hashOps = redisTemplate.opsForHash();
        String key = aggregateKey(transactionId);
        hashOps.put(key, FIELD_RULE_SCORE, String.valueOf(event.getRuleScore()));
        storeIfPresent(hashOps, key, FIELD_RULE_BAND, event.getRuleBand());
        storeIfPresent(hashOps, key, FIELD_RULE_MATCHES, serializeList(event.getRuleMatches()));
        storeIfPresent(hashOps, key, FIELD_RULE_VERSION, event.getRuleVersion());
        storeIfPresent(hashOps, key, FIELD_FEATURES, serializeMap(event.getFeatures()));
        storeIfPresent(hashOps, key, FIELD_SENDER_USER_ID, event.getSenderUserId());
        storeIfPresent(hashOps, key, FIELD_RECEIVER_USER_ID, event.getReceiverUserId());
        storeIfPresent(hashOps, key, FIELD_MERCHANT_ID, event.getMerchantId());
        storeIfPresent(hashOps, key, FIELD_AMOUNT, event.getAmount());
        storeIfPresent(hashOps, key, FIELD_CURRENCY, event.getCurrency());
        redisTemplate.expire(key, AGGREGATION_TTL);

        if ("SAFE".equalsIgnoreCase(event.getRuleBand())) {
            finalizeDecision(transactionId, FinalDecisionInput.fromRule(event, "RULE_SAFE", FraudDecision.ALLOW));
        } else if ("RISK".equalsIgnoreCase(event.getRuleBand())) {
            finalizeDecision(transactionId, FinalDecisionInput.fromRule(event, "RULE_RISK", FraudDecision.BLOCK));
        }
    }

    public void handleMlEvent(MlScoreEvent event) {
        String transactionId = event.getTransactionId();
        if (transactionId == null) {
            return;
        }

        HashOperations<String, Object, Object> hashOps = redisTemplate.opsForHash();
        String key = aggregateKey(transactionId);
        hashOps.put(key, FIELD_ML_SCORE, String.valueOf(event.getMlScore()));
        storeIfPresent(hashOps, key, FIELD_MODEL_VERSION, event.getModelVersion());
        redisTemplate.expire(key, AGGREGATION_TTL);

        Map<Object, Object> values = hashOps.entries(key);
        String band = values.get(FIELD_RULE_BAND) != null ? values.get(FIELD_RULE_BAND).toString() : null;
        if (!"GRAY".equalsIgnoreCase(band)) {
            return;
        }

        double mlScore = parseDouble(values.get(FIELD_ML_SCORE));
        String mlBand = mlScore < 0.30 ? "SAFE" : (mlScore > 0.70 ? "RISK" : "GRAY");
        FraudDecision decision = mlScore < 0.30
                ? FraudDecision.ALLOW
                : (mlScore > 0.70 ? FraudDecision.BLOCK : FraudDecision.HOLD);
        String reason = mlScore < 0.30
                ? "ML_SAFE"
                : (mlScore > 0.70 ? "ML_RISK" : "ML_GRAY");

        finalizeDecision(transactionId, FinalDecisionInput.fromMl(decision, reason, mlBand));
    }

    private void finalizeDecision(String transactionId, FinalDecisionInput input) {
        String finalizeKey = "fraud:finalized:" + transactionId;
        Boolean first = redisTemplate.opsForValue()
                .setIfAbsent(finalizeKey, "1", Duration.ofMinutes(2));
        if (first == null || !first) {
            return;
        }

        Map<Object, Object> values = redisTemplate.opsForHash().entries(aggregateKey(transactionId));
        Map<String, Object> features = deserializeMap(values.get(FIELD_FEATURES));

        Long senderUserId = parseLong(values.get(FIELD_SENDER_USER_ID), features, "senderUserId");
        Long receiverUserId = parseLong(values.get(FIELD_RECEIVER_USER_ID), features, "receiverUserId");
        Long merchantId = parseLong(values.get(FIELD_MERCHANT_ID), features, "merchantId");
        BigDecimal amount = parseBigDecimal(values.get(FIELD_AMOUNT), features, "amount");
        String currency = parseString(values.get(FIELD_CURRENCY), features, "currency");
        String country = parseString(values.get(FIELD_COUNTRY), features, "senderAccountCountry");

        boolean blacklistHit = input.blacklistHit != null ? input.blacklistHit : parseBoolean(values.get(FIELD_BLACKLIST_HIT));
        Double ruleScore = input.ruleScore != null ? input.ruleScore : parseDouble(values.get(FIELD_RULE_SCORE));
        String ruleBand = input.ruleBand != null ? input.ruleBand : parseString(values.get(FIELD_RULE_BAND), null, null);
        List<String> ruleMatches = input.ruleMatches != null ? input.ruleMatches : deserializeList(values.get(FIELD_RULE_MATCHES));
        Double mlScore = input.mlScore != null ? input.mlScore : parseDouble(values.get(FIELD_ML_SCORE));
        String mlBand = input.mlBand;
        String modelVersion = input.modelVersion != null ? input.modelVersion : parseString(values.get(FIELD_MODEL_VERSION), null, null);
        Integer ruleVersion = input.ruleVersion != null ? input.ruleVersion : parseInt(values.get(FIELD_RULE_VERSION));

        FraudFinalDecisionEvent out = new FraudFinalDecisionEvent(
                transactionId,
                senderUserId,
                senderUserId,
                receiverUserId,
                merchantId,
                amount,
                currency,
                country,
                serializeMap(features),
                blacklistHit,
                ruleScore,
                ruleBand,
                ruleMatches,
                mlScore,
                mlBand,
                input.finalDecision,
                input.decisionReason,
                modelVersion,
                ruleVersion,
                Instant.now()
        );

        persistDecision(out);
        kafkaTemplate.send(OUT_TOPIC, transactionId, out);
        redisTemplate.delete(aggregateKey(transactionId));
    }

    private void persistDecision(FraudFinalDecisionEvent event) {
        FraudDecisionRecord record = new FraudDecisionRecord();
        record.setTransactionId(event.getTransactionId());
        record.setAccountId(event.getAccountId());
        record.setAmount(event.getAmount());
        record.setCountry(event.getCountry());
        record.setFeaturesJson(event.getFeaturesJson());
        record.setBlacklistHit(event.isBlacklistHit());
        record.setRuleScore(event.getRuleScore());
        record.setRuleBand(event.getRuleBand());
        record.setRuleMatches(serializeList(event.getRuleMatches()));
        record.setMlScore(event.getMlScore());
        record.setMlBand(event.getMlBand());
        record.setFinalDecision(event.getFinalDecision() != null ? event.getFinalDecision().name() : null);
        record.setDecisionReason(event.getDecisionReason());
        record.setModelVersion(event.getModelVersion());
        record.setRuleVersion(event.getRuleVersion());
        record.setCreatedAt(event.getDecidedAt() != null ? event.getDecidedAt() : Instant.now());
        decisionRepository.save(record);
    }

    private void storeTransactionSnapshot(String transactionId, Map<String, Object> transaction) {
        HashOperations<String, Object, Object> hashOps = redisTemplate.opsForHash();
        String key = aggregateKey(transactionId);
        storeIfPresent(hashOps, key, FIELD_FEATURES, serializeMap(transaction));
        storeIfPresent(hashOps, key, FIELD_SENDER_USER_ID, transaction.get("senderUserId"));
        storeIfPresent(hashOps, key, FIELD_RECEIVER_USER_ID, transaction.get("receiverUserId"));
        storeIfPresent(hashOps, key, FIELD_MERCHANT_ID, transaction.get("merchantId"));
        storeIfPresent(hashOps, key, FIELD_AMOUNT, transaction.get("amount"));
        storeIfPresent(hashOps, key, FIELD_CURRENCY, transaction.get("currency"));
        storeIfPresent(hashOps, key, FIELD_COUNTRY, transaction.get("senderAccountCountry"));
        redisTemplate.expire(key, AGGREGATION_TTL);
    }

    private String serializeList(List<String> values) {
        try {
            return objectMapper.writeValueAsString(values == null ? List.of() : values);
        } catch (Exception e) {
            return "[]";
        }
    }

    private String serializeMap(Map<String, Object> values) {
        try {
            return objectMapper.writeValueAsString(values == null ? Collections.emptyMap() : values);
        } catch (Exception e) {
            return "{}";
        }
    }

    private List<String> deserializeList(Object raw) {
        if (raw == null) {
            return List.of();
        }
        try {
            return objectMapper.readValue(raw.toString(), new TypeReference<List<String>>() {});
        } catch (Exception e) {
            return List.of();
        }
    }

    private Map<String, Object> deserializeMap(Object raw) {
        if (raw == null) {
            return Collections.emptyMap();
        }
        try {
            return objectMapper.readValue(raw.toString(), new TypeReference<Map<String, Object>>() {});
        } catch (Exception e) {
            return Collections.emptyMap();
        }
    }

    private double parseDouble(Object value) {
        if (value == null) {
            return 0.0;
        }
        try {
            return Double.parseDouble(value.toString());
        } catch (NumberFormatException e) {
            return 0.0;
        }
    }

    private Long parseLong(Object value, Map<String, Object> features, String fallbackKey) {
        if (value != null) {
            try {
                return Long.parseLong(value.toString());
            } catch (NumberFormatException ignored) {
            }
        }
        Object fallback = features != null ? features.get(fallbackKey) : null;
        if (fallback == null) {
            return null;
        }
        try {
            return Long.parseLong(fallback.toString());
        } catch (NumberFormatException e) {
            return null;
        }
    }

    private BigDecimal parseBigDecimal(Object value, Map<String, Object> features, String fallbackKey) {
        Object raw = value;
        if (raw == null && features != null) {
            raw = features.get(fallbackKey);
        }
        if (raw == null) {
            return null;
        }
        try {
            return new BigDecimal(raw.toString());
        } catch (NumberFormatException e) {
            return null;
        }
    }

    private String parseString(Object value, Map<String, Object> features, String fallbackKey) {
        Object raw = value;
        if (raw == null && features != null) {
            raw = features.get(fallbackKey);
        }
        return raw != null ? raw.toString() : null;
    }

    private Integer parseInt(Object value) {
        if (value == null) {
            return null;
        }
        try {
            return Integer.parseInt(value.toString());
        } catch (NumberFormatException e) {
            return null;
        }
    }

    private boolean parseBoolean(Object value) {
        if (value == null) {
            return false;
        }
        return Boolean.parseBoolean(value.toString());
    }

    private void storeIfPresent(
            HashOperations<String, Object, Object> hashOps,
            String key,
            String field,
            Object value
    ) {
        if (value != null) {
            hashOps.put(key, field, value.toString());
        }
    }

    private String aggregateKey(String transactionId) {
        return "fraud:aggregate:" + transactionId;
    }

    private static final class FinalDecisionInput {
        private final FraudDecision finalDecision;
        private final String decisionReason;
        private final Boolean blacklistHit;
        private final Double ruleScore;
        private final String ruleBand;
        private final List<String> ruleMatches;
        private final Double mlScore;
        private final String mlBand;
        private final String modelVersion;
        private final Integer ruleVersion;

        private FinalDecisionInput(
                FraudDecision finalDecision,
                String decisionReason,
                Boolean blacklistHit,
                Double ruleScore,
                String ruleBand,
                List<String> ruleMatches,
                Double mlScore,
                String mlBand,
                String modelVersion,
                Integer ruleVersion
        ) {
            this.finalDecision = finalDecision;
            this.decisionReason = decisionReason;
            this.blacklistHit = blacklistHit;
            this.ruleScore = ruleScore;
            this.ruleBand = ruleBand;
            this.ruleMatches = ruleMatches;
            this.mlScore = mlScore;
            this.mlBand = mlBand;
            this.modelVersion = modelVersion;
            this.ruleVersion = ruleVersion;
        }

        static FinalDecisionInput fromBlacklist(BlacklistCheckEvent event) {
            return new FinalDecisionInput(
                    FraudDecision.BLOCK,
                    "BLACKLIST",
                    true,
                    null,
                    null,
                    null,
                    null,
                    null,
                    null,
                    null
            );
        }

        static FinalDecisionInput fromRule(RuleEvaluationEvent event, String reason, FraudDecision decision) {
            return new FinalDecisionInput(
                    decision,
                    reason,
                    false,
                    event.getRuleScore(),
                    event.getRuleBand(),
                    event.getRuleMatches(),
                    null,
                    null,
                    null,
                    event.getRuleVersion()
            );
        }

        static FinalDecisionInput fromMl(FraudDecision decision, String reason, String mlBand) {
            return new FinalDecisionInput(
                    decision,
                    reason,
                    false,
                    null,
                    "GRAY",
                    null,
                    null,
                    mlBand,
                    null,
                    null
            );
        }
    }
}
