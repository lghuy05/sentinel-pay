package com.example.blacklist_service.messaging;

import com.example.blacklist_service.event.BlacklistCheckEvent;
import com.example.blacklist_service.event.TransactionEnrichedEvent;
import com.example.blacklist_service.service.BlacklistCheckService;
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
public class TransactionEnrichedListener {

    private static final Logger log = LoggerFactory.getLogger(TransactionEnrichedListener.class);
    private static final String OUT_TOPIC = "fraud.blacklist";
    private static final AtomicLong LOG_COUNTER = new AtomicLong(0);

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
    public void handleEnriched(ConsumerRecord<String, String> record) throws Exception {
        String payload = record.value();
        boolean logSample = shouldLog();
        if (logSample) {
            long lagMs = record.timestamp() > 0 ? System.currentTimeMillis() - record.timestamp() : -1;
            log.info("Kafka consume topic=transactions.enriched partition={} offset={} lagMs={}",
                    record.partition(),
                    record.offset(),
                    lagMs);
        }
        TransactionEnrichedEvent event = objectMapper.readValue(payload, TransactionEnrichedEvent.class);
        BlacklistCheckEvent check = blacklistCheckService.check(event);

        long startNs = System.nanoTime();
        RecordMetadata metadata = kafkaTemplate.send(OUT_TOPIC, check.getTransactionId(), check)
                .get()
                .getRecordMetadata();

        if (logSample) {
            long ackMs = (System.nanoTime() - startNs) / 1_000_000;
            log.info(
                    "Kafka published BlacklistCheckEvent txId={} partition={} offset={} ackMs={}",
                    check.getTransactionId(),
                    metadata.partition(),
                    metadata.offset(),
                    ackMs
            );
            log.info(
                    "Blacklist check txId={} hit={} reason={}",
                    check.getTransactionId(),
                    check.isBlacklistHit(),
                    check.getReason()
            );
        }
    }

    private boolean shouldLog() {
        return LOG_COUNTER.incrementAndGet() % 100 == 0;
    }
}
