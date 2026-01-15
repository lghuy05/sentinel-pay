package com.example.fraud_orchestrator.messaging;

import com.example.fraud_orchestrator.event.BlacklistCheckEvent;
import com.example.fraud_orchestrator.event.MlScoreEvent;
import com.example.fraud_orchestrator.event.RuleEvaluationEvent;
import com.example.fraud_orchestrator.service.FraudOrchestrationService;
import com.fasterxml.jackson.databind.ObjectMapper;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

@Component
public class FraudSignalListener {

    private static final Logger log = LoggerFactory.getLogger(FraudSignalListener.class);

    private final FraudOrchestrationService orchestrationService;
    private final ObjectMapper objectMapper;

    public FraudSignalListener(
            FraudOrchestrationService orchestrationService,
            ObjectMapper objectMapper
    ) {
        this.orchestrationService = orchestrationService;
        this.objectMapper = objectMapper;
    }

    @KafkaListener(topics = "fraud.rules", groupId = "fraud-orchestrator")
    public void handleRuleSignal(String payload) {
        try {
            RuleEvaluationEvent event =
                    objectMapper.readValue(payload, RuleEvaluationEvent.class);
            orchestrationService.handleRuleEvent(event);
        } catch (Exception e) {
            log.error("Failed to parse rule signal payload={}", payload, e);
        }
    }

    @KafkaListener(topics = "fraud.blacklist", groupId = "fraud-orchestrator")
    public void handleBlacklistSignal(String payload) {
        try {
            BlacklistCheckEvent event =
                    objectMapper.readValue(payload, BlacklistCheckEvent.class);
            orchestrationService.handleBlacklistEvent(event);
        } catch (Exception e) {
            log.error("Failed to parse blacklist signal payload={}", payload, e);
        }
    }

    @KafkaListener(topics = "fraud.ml", groupId = "fraud-orchestrator")
    public void handleMlSignal(String payload) {
        try {
            MlScoreEvent event =
                    objectMapper.readValue(payload, MlScoreEvent.class);
            orchestrationService.handleMlEvent(event);
        } catch (Exception e) {
            log.error("Failed to parse ml signal payload={}", payload, e);
        }
    }
}
