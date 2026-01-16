package com.example.fraud_orchestrator.service;

import java.time.Duration;
import java.time.Instant;
import java.util.Collections;
import java.util.List;
import java.util.Map;

import org.springframework.data.redis.core.HashOperations;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

import com.example.fraud_orchestrator.event.BlacklistCheckEvent;
import com.example.fraud_orchestrator.event.FraudDecision;
import com.example.fraud_orchestrator.event.FraudFinalDecisionEvent;
import com.example.fraud_orchestrator.event.MlScoreEvent;
import com.example.fraud_orchestrator.event.RuleEvaluationEvent;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;

@Service
public class FraudOrchestrationService {

    private static final String OUT_TOPIC = "fraud.final";
    private static final Duration AGGREGATION_TTL = Duration.ofSeconds(5);

    private static final String FIELD_RULE_SCORE = "ruleScore";
    private static final String FIELD_RULE_MATCHES = "ruleMatches";
    private static final String FIELD_BLACKLIST_SCORE = "blacklistScore";
    private static final String FIELD_BLACKLIST_MATCHES = "blacklistMatches";
    private static final String FIELD_ML_SCORE = "mlScore";
    private static final String FIELD_SENDER_USER_ID = "senderUserId";
    private static final String FIELD_RECEIVER_USER_ID = "receiverUserId";
    private static final String FIELD_MERCHANT_ID = "merchantId";
    private static final String FIELD_AMOUNT = "amount";
    private static final String FIELD_CURRENCY = "currency";
    private static final String FIELD_RISK_SCORE = "riskScore";
    private static final String FIELD_RISK_LEVEL = "riskLevel";
    private static final String FIELD_TRIGGERED_RULES = "triggeredRules";
    private static final String FIELD_HARD_STOP_MATCHES = "hardStopMatches";
    private static final String FIELD_HARD_STOP_DECISION = "hardStopDecision";

    private final StringRedisTemplate redisTemplate;
    private final KafkaTemplate<String, FraudFinalDecisionEvent> kafkaTemplate;
    private final ObjectMapper objectMapper;

    public FraudOrchestrationService(
            StringRedisTemplate redisTemplate,
            KafkaTemplate<String, FraudFinalDecisionEvent> kafkaTemplate,
            ObjectMapper objectMapper
    ) {
        this.redisTemplate = redisTemplate;
        this.kafkaTemplate = kafkaTemplate;
        this.objectMapper = objectMapper;
    }

    public void handleRuleEvent(RuleEvaluationEvent event) {
        String key = aggregateKey(event.getTransactionId());
        HashOperations<String, Object, Object> hashOps = redisTemplate.opsForHash();
        hashOps.put(key, FIELD_RULE_SCORE, String.valueOf(event.getRuleScore()));
        hashOps.put(key, FIELD_RULE_MATCHES, serializeList(event.getMatchedRules()));
        hashOps.put(key, FIELD_RISK_SCORE, String.valueOf(event.getRiskScore()));
        hashOps.put(key, FIELD_RISK_LEVEL, event.getRiskLevel());
        hashOps.put(key, FIELD_TRIGGERED_RULES, serializeList(event.getTriggeredRules()));
        hashOps.put(key, FIELD_HARD_STOP_MATCHES, serializeList(event.getHardStopMatches()));
        storeIfPresent(hashOps, key, FIELD_HARD_STOP_DECISION, event.getHardStopDecision());
        storeIfPresent(hashOps, key, FIELD_SENDER_USER_ID, event.getSenderUserId());
        storeIfPresent(hashOps, key, FIELD_RECEIVER_USER_ID, event.getReceiverUserId());
        storeIfPresent(hashOps, key, FIELD_MERCHANT_ID, event.getMerchantId());
        storeIfPresent(hashOps, key, FIELD_AMOUNT, event.getAmount());
        storeIfPresent(hashOps, key, FIELD_CURRENCY, event.getCurrency());
        redisTemplate.expire(key, AGGREGATION_TTL);
        tryFinalize(event.getTransactionId());
    }

    public void handleBlacklistEvent(BlacklistCheckEvent event) {
        String key = aggregateKey(event.getTransactionId());
        HashOperations<String, Object, Object> hashOps = redisTemplate.opsForHash();
        hashOps.put(key, FIELD_BLACKLIST_SCORE, String.valueOf(event.getBlacklistScore()));
        hashOps.put(key, FIELD_BLACKLIST_MATCHES, serializeList(event.getMatchedEntries()));
        redisTemplate.expire(key, AGGREGATION_TTL);
        tryFinalize(event.getTransactionId());
    }

    public void handleMlEvent(MlScoreEvent event) {
        String key = aggregateKey(event.getTransactionId());
        HashOperations<String, Object, Object> hashOps = redisTemplate.opsForHash();
        hashOps.put(key, FIELD_ML_SCORE, String.valueOf(event.getMlScore()));
        redisTemplate.expire(key, AGGREGATION_TTL);
        tryFinalize(event.getTransactionId());
    }

    private void tryFinalize(String transactionId) {
        String key = aggregateKey(transactionId);
        Map<Object, Object> values = redisTemplate.opsForHash().entries(key);
        if (!values.keySet().containsAll(
                List.of(FIELD_RULE_SCORE, FIELD_BLACKLIST_SCORE, FIELD_ML_SCORE)
        )) {
            return;
        }

        String finalizeKey = "fraud:finalized:" + transactionId;
        Boolean first = redisTemplate.opsForValue()
                .setIfAbsent(finalizeKey, "1", Duration.ofMinutes(1));
        if (first == null || !first) {
            return;
        }

        double ruleScore = parseDouble(values.get(FIELD_RULE_SCORE));
        double blacklistScore = parseDouble(values.get(FIELD_BLACKLIST_SCORE));
        double mlScore = parseDouble(values.get(FIELD_ML_SCORE));

        double finalScore = (ruleScore * 0.5) + (blacklistScore * 0.3) + (mlScore * 0.2);
        List<String> ruleMatches = deserializeList(values.get(FIELD_RULE_MATCHES));
        List<String> blacklistMatches = deserializeList(values.get(FIELD_BLACKLIST_MATCHES));
        int riskScore = (int) parseDouble(values.get(FIELD_RISK_SCORE));
        String riskLevel = values.get(FIELD_RISK_LEVEL) != null
                ? values.get(FIELD_RISK_LEVEL).toString()
                : null;
        List<String> triggeredRules = deserializeList(values.get(FIELD_TRIGGERED_RULES));
        List<String> hardStopMatches = deserializeList(values.get(FIELD_HARD_STOP_MATCHES));
        String hardStopDecision = values.get(FIELD_HARD_STOP_DECISION) != null
                ? values.get(FIELD_HARD_STOP_DECISION).toString()
                : null;
        Long senderUserId = parseLong(values.get(FIELD_SENDER_USER_ID));
        Long receiverUserId = parseLong(values.get(FIELD_RECEIVER_USER_ID));
        Long merchantId = parseLong(values.get(FIELD_MERCHANT_ID));
        java.math.BigDecimal amount = parseBigDecimal(values.get(FIELD_AMOUNT));
        String currency = values.get(FIELD_CURRENCY) != null
                ? values.get(FIELD_CURRENCY).toString()
                : null;

        FraudDecision decision = decide(finalScore, blacklistScore, ruleMatches, hardStopDecision);

        FraudFinalDecisionEvent out = new FraudFinalDecisionEvent(
                transactionId,
                senderUserId,
                receiverUserId,
                merchantId,
                amount,
                currency,
                decision,
                ruleScore,
                blacklistScore,
                mlScore,
                finalScore,
                ruleMatches,
                blacklistMatches,
                riskScore,
                riskLevel,
                triggeredRules,
                hardStopMatches,
                hardStopDecision,
                Instant.now()
        );

        kafkaTemplate.send(OUT_TOPIC, transactionId, out);
        redisTemplate.delete(key);
    }

    private FraudDecision decide(
            double finalScore,
            double blacklistScore,
            List<String> ruleMatches,
            String hardStopDecision
    ) {
        if ("BLOCK".equalsIgnoreCase(hardStopDecision)) {
            return FraudDecision.BLOCK;
        }
        if (ruleMatches.contains("absolute-high-amount")) {
            return FraudDecision.BLOCK;
        }
        if (blacklistScore >= 1.0) {
            return FraudDecision.BLOCK;
        }
        if ("REVIEW".equalsIgnoreCase(hardStopDecision)) {
            return FraudDecision.HOLD;
        }
        if (finalScore >= 0.8) {
            return FraudDecision.BLOCK;
        }
        if (finalScore >= 0.5) {
            return FraudDecision.HOLD;
        }
        return FraudDecision.ALLOW;
    }

    private String serializeList(List<String> values) {
        try {
            return objectMapper.writeValueAsString(values == null ? List.of() : values);
        } catch (Exception e) {
            return "[]";
        }
    }

    private List<String> deserializeList(Object raw) {
        if (raw == null) {
            return Collections.emptyList();
        }
        try {
            return objectMapper.readValue(raw.toString(), new TypeReference<List<String>>() {});
        } catch (Exception e) {
            return Collections.emptyList();
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

    private Long parseLong(Object value) {
        if (value == null) {
            return null;
        }
        try {
            return Long.parseLong(value.toString());
        } catch (NumberFormatException e) {
            return null;
        }
    }

    private java.math.BigDecimal parseBigDecimal(Object value) {
        if (value == null) {
            return null;
        }
        try {
            return new java.math.BigDecimal(value.toString());
        } catch (NumberFormatException e) {
            return null;
        }
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
}
