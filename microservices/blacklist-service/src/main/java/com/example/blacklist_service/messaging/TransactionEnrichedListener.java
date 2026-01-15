package com.example.blacklist_service.messaging;

import com.example.blacklist_service.event.BlacklistCheckEvent;
import com.example.blacklist_service.event.TransactionEnrichedEvent;
import com.example.blacklist_service.service.BlacklistCheckService;
import com.fasterxml.jackson.databind.ObjectMapper;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

@Component
public class TransactionEnrichedListener {

    private static final Logger log = LoggerFactory.getLogger(TransactionEnrichedListener.class);
    private static final String OUT_TOPIC = "fraud.blacklist";

    private final BlacklistCheckService blacklistCheckService;
    private final KafkaTemplate<String, BlacklistCheckEvent> kafkaTemplate;
    private final ObjectMapper objectMapper;

    public TransactionEnrichedListener(
            BlacklistCheckService blacklistCheckService,
            KafkaTemplate<String, BlacklistCheckEvent> kafkaTemplate,
            ObjectMapper objectMapper
    ) {
        this.blacklistCheckService = blacklistCheckService;
        this.kafkaTemplate = kafkaTemplate;
        this.objectMapper = objectMapper;
    }

    @KafkaListener(topics = "transactions.enriched", groupId = "blacklist-service")
    public void handleEnriched(String payload) {
        try {
            TransactionEnrichedEvent event =
                    objectMapper.readValue(payload, TransactionEnrichedEvent.class);
            BlacklistCheckEvent check = blacklistCheckService.check(event);
            kafkaTemplate.send(OUT_TOPIC, check.getTransactionId(), check);
            log.info("Blacklist check txId={} score={}", check.getTransactionId(), check.getBlacklistScore());
        } catch (Exception e) {
            log.error("Failed to evaluate blacklist payload={}", payload, e);
        }
    }
}
