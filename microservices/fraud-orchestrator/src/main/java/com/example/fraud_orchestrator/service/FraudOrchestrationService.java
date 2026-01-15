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
        FraudDecision decision = decide(finalScore, blacklistScore);

        List<String> ruleMatches = deserializeList(values.get(FIELD_RULE_MATCHES));
        List<String> blacklistMatches = deserializeList(values.get(FIELD_BLACKLIST_MATCHES));

        FraudFinalDecisionEvent out = new FraudFinalDecisionEvent(
                transactionId,
                decision,
                ruleScore,
                blacklistScore,
                mlScore,
                finalScore,
                ruleMatches,
                blacklistMatches,
                Instant.now()
        );

        kafkaTemplate.send(OUT_TOPIC, transactionId, out);
        redisTemplate.delete(key);
    }

    private FraudDecision decide(double finalScore, double blacklistScore) {
        if (blacklistScore >= 1.0) {
            return FraudDecision.BLOCK;
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

    private String aggregateKey(String transactionId) {
        return "fraud:aggregate:" + transactionId;
    }
}
