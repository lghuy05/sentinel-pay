package com.example.alert_service.messaging;

import com.example.alert_service.event.FraudFinalDecisionEvent;
import com.example.alert_service.service.FraudDecisionService;
import com.fasterxml.jackson.databind.ObjectMapper;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

@Component
public class FraudFinalDecisionListener {

    private static final Logger log = LoggerFactory.getLogger(FraudFinalDecisionListener.class);

    private final FraudDecisionService fraudDecisionService;
    private final ObjectMapper objectMapper;

    public FraudFinalDecisionListener(
            FraudDecisionService fraudDecisionService,
            ObjectMapper objectMapper
    ) {
        this.fraudDecisionService = fraudDecisionService;
        this.objectMapper = objectMapper;
    }

    @KafkaListener(topics = "fraud.final", groupId = "alert-service")
    public void handleFinalDecision(String payload) {
        try {
            FraudFinalDecisionEvent event =
                    objectMapper.readValue(payload, FraudFinalDecisionEvent.class);
            fraudDecisionService.handleDecision(event);
        } catch (Exception e) {
            log.error("Failed to handle final decision payload={}", payload, e);
        }
    }
}
