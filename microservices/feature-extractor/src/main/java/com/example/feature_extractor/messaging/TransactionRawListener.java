package com.example.feature_extractor.messaging;

import com.example.feature_extractor.event.TransactionEnrichedEvent;
import com.example.feature_extractor.event.TransactionReceivedEvent;
import com.example.feature_extractor.service.FeatureExtractionService;
import com.fasterxml.jackson.databind.ObjectMapper;

import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.apache.kafka.clients.producer.RecordMetadata;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

import java.util.concurrent.atomic.AtomicLong;

@Component
public class TransactionRawListener {

    private static final Logger log = LoggerFactory.getLogger(TransactionRawListener.class);
    private static final String OUT_TOPIC = "transactions.enriched";
    private static final AtomicLong LOG_COUNTER = new AtomicLong(0);

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
    public void handleRawTransaction(ConsumerRecord<String, String> record) throws Exception {
        String payload = record.value();
        boolean logSample = shouldLog();
        if (logSample) {
            long lagMs = record.timestamp() > 0 ? System.currentTimeMillis() - record.timestamp() : -1;
            log.info("Kafka consume topic=transactions.raw partition={} offset={} lagMs={}",
                    record.partition(),
                    record.offset(),
                    lagMs);
        }
        TransactionReceivedEvent event = objectMapper.readValue(payload, TransactionReceivedEvent.class);

        TransactionEnrichedEvent enriched = featureExtractionService.extract(event);
        long startNs = System.nanoTime();
        RecordMetadata metadata = kafkaTemplate.send(OUT_TOPIC, enriched.getTransactionId(), enriched)
                .get()
                .getRecordMetadata();

        if (logSample) {
            long ackMs = (System.nanoTime() - startNs) / 1_000_000;
            log.info(
                    "Kafka published TransactionEnrichedEvent txId={} partition={} offset={} ackMs={}",
                    enriched.getTransactionId(),
                    metadata.partition(),
                    metadata.offset(),
                    ackMs
            );
        }
    }

    private boolean shouldLog() {
        return LOG_COUNTER.incrementAndGet() % 100 == 0;
    }
}
