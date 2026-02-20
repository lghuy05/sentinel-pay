package com.example.rule_engine.messaging;

import com.example.rule_engine.event.BlacklistCheckEvent;
import com.example.rule_engine.event.RuleEvaluationEvent;
import com.example.rule_engine.event.TransactionEnrichedEvent;
import com.example.rule_engine.service.RuleEngineService;
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
    private static final String OUT_TOPIC = "fraud.rules";
    private static final AtomicLong LOG_COUNTER = new AtomicLong(0);

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

    @KafkaListener(topics = "fraud.blacklist", groupId = "rule-engine")
    public void handleBlacklist(ConsumerRecord<String, String> record) throws Exception {
        String payload = record.value();
        boolean logSample = shouldLog();
        if (logSample) {
            long lagMs = record.timestamp() > 0 ? System.currentTimeMillis() - record.timestamp() : -1;
            log.info("Kafka consume topic=fraud.blacklist partition={} offset={} lagMs={}",
                    record.partition(),
                    record.offset(),
                    lagMs);
        }
        BlacklistCheckEvent event = objectMapper.readValue(payload, BlacklistCheckEvent.class);
        if (event.isBlacklistHit()) {
            if (logSample) {
                log.info("Skipping rules for blacklisted txId={} reason={}", event.getTransactionId(), event.getReason());
            }
            return;
        }
        TransactionEnrichedEvent enriched = event.getTransaction();
        if (enriched == null) {
            log.warn("Missing transaction payload for txId={}", event.getTransactionId());
            return;
        }
        RuleEvaluationEvent evaluation = ruleEngineService.evaluate(enriched);

        long startNs = System.nanoTime();
        RecordMetadata metadata = kafkaTemplate.send(OUT_TOPIC, evaluation.getTransactionId(), evaluation)
                .get()
                .getRecordMetadata();

        if (logSample) {
            long ackMs = (System.nanoTime() - startNs) / 1_000_000;
            log.info(
                    "Kafka published RuleEvaluationEvent txId={} partition={} offset={} ackMs={}",
                    evaluation.getTransactionId(),
                    metadata.partition(),
                    metadata.offset(),
                    ackMs
            );
            log.info("Rule evaluation txId={} score={} band={}", evaluation.getTransactionId(), evaluation.getRuleScore(), evaluation.getRuleBand());
        }
    }

    private boolean shouldLog() {
        return LOG_COUNTER.incrementAndGet() % 100 == 0;
    }
}
