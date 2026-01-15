package com.example.rule_engine.messaging;

import com.example.rule_engine.event.RuleEvaluationEvent;
import com.example.rule_engine.event.TransactionEnrichedEvent;
import com.example.rule_engine.service.RuleEngineService;
import com.fasterxml.jackson.databind.ObjectMapper;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

@Component
public class TransactionEnrichedListener {

    private static final Logger log = LoggerFactory.getLogger(TransactionEnrichedListener.class);
    private static final String OUT_TOPIC = "fraud.rules";

    private final RuleEngineService ruleEngineService;
    private final KafkaTemplate<String, RuleEvaluationEvent> kafkaTemplate;
    private final ObjectMapper objectMapper;

    public TransactionEnrichedListener(
            RuleEngineService ruleEngineService,
            KafkaTemplate<String, RuleEvaluationEvent> kafkaTemplate,
            ObjectMapper objectMapper
    ) {
        this.ruleEngineService = ruleEngineService;
        this.kafkaTemplate = kafkaTemplate;
        this.objectMapper = objectMapper;
    }

    @KafkaListener(topics = "transactions.enriched", groupId = "rule-engine")
    public void handleEnriched(String payload) {
        try {
            TransactionEnrichedEvent event =
                    objectMapper.readValue(payload, TransactionEnrichedEvent.class);
            RuleEvaluationEvent evaluation = ruleEngineService.evaluate(event);
            kafkaTemplate.send(OUT_TOPIC, evaluation.getTransactionId(), evaluation);
            log.info("Rule evaluation txId={} score={}", evaluation.getTransactionId(), evaluation.getRuleScore());
        } catch (Exception e) {
            log.error("Failed to evaluate rules payload={}", payload, e);
        }
    }
}
