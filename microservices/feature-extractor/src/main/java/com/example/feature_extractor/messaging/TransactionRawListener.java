package com.example.feature_extractor.messaging;

import com.example.feature_extractor.event.TransactionEnrichedEvent;
import com.example.feature_extractor.event.TransactionReceivedEvent;
import com.example.feature_extractor.service.FeatureExtractionService;
import com.fasterxml.jackson.databind.ObjectMapper;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

@Component
public class TransactionRawListener {

    private static final Logger log = LoggerFactory.getLogger(TransactionRawListener.class);
    private static final String OUT_TOPIC = "transactions.enriched";

    private final FeatureExtractionService featureExtractionService;
    private final KafkaTemplate<String, TransactionEnrichedEvent> kafkaTemplate;
    private final ObjectMapper objectMapper;

    public TransactionRawListener(
            FeatureExtractionService featureExtractionService,
            KafkaTemplate<String, TransactionEnrichedEvent> kafkaTemplate,
            ObjectMapper objectMapper
    ) {
        this.featureExtractionService = featureExtractionService;
        this.kafkaTemplate = kafkaTemplate;
        this.objectMapper = objectMapper;
    }

    @KafkaListener(topics = "transactions.raw", groupId = "feature-extractor")
    public void handleRawTransaction(String payload) {
        try {
            TransactionReceivedEvent event =
                    objectMapper.readValue(payload, TransactionReceivedEvent.class);

            TransactionEnrichedEvent enriched = featureExtractionService.extract(event);
            kafkaTemplate.send(OUT_TOPIC, enriched.getTransactionId(), enriched);
            log.info("Enriched transaction txId={}", enriched.getTransactionId());
        } catch (Exception e) {
            log.error("Failed to enrich transaction payload={}", payload, e);
        }
    }
}
