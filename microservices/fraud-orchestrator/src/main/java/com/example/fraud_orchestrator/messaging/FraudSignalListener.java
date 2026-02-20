package com.example.fraud_orchestrator.messaging;

import com.example.fraud_orchestrator.event.BlacklistCheckEvent;
import com.example.fraud_orchestrator.event.MlScoreEvent;
import com.example.fraud_orchestrator.event.RuleEvaluationEvent;
import com.example.fraud_orchestrator.service.FraudOrchestrationService;
import com.fasterxml.jackson.databind.ObjectMapper;

import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

import java.util.concurrent.atomic.AtomicLong;

@Component
public class FraudSignalListener {

    private static final Logger log = LoggerFactory.getLogger(FraudSignalListener.class);
    private static final AtomicLong LOG_COUNTER = new AtomicLong(0);

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
    public void handleRuleSignal(ConsumerRecord<String, String> record) throws Exception {
        String payload = record.value();
        if (shouldLog()) {
            long lagMs = record.timestamp() > 0 ? System.currentTimeMillis() - record.timestamp() : -1;
            log.info("Kafka consume topic=fraud.rules partition={} offset={} lagMs={}",
                    record.partition(),
                    record.offset(),
                    lagMs);
        }
        RuleEvaluationEvent event =
                objectMapper.readValue(payload, RuleEvaluationEvent.class);
        orchestrationService.handleRuleEvent(event);
    }

    @KafkaListener(topics = "fraud.blacklist", groupId = "fraud-orchestrator")
    public void handleBlacklistSignal(ConsumerRecord<String, String> record) throws Exception {
        String payload = record.value();
        if (shouldLog()) {
            long lagMs = record.timestamp() > 0 ? System.currentTimeMillis() - record.timestamp() : -1;
            log.info("Kafka consume topic=fraud.blacklist partition={} offset={} lagMs={}",
                    record.partition(),
                    record.offset(),
                    lagMs);
        }
        BlacklistCheckEvent event =
                objectMapper.readValue(payload, BlacklistCheckEvent.class);
        orchestrationService.handleBlacklistEvent(event);
    }

    @KafkaListener(topics = "fraud.ml", groupId = "fraud-orchestrator")
    public void handleMlSignal(ConsumerRecord<String, String> record) throws Exception {
        String payload = record.value();
        if (shouldLog()) {
            long lagMs = record.timestamp() > 0 ? System.currentTimeMillis() - record.timestamp() : -1;
            log.info("Kafka consume topic=fraud.ml partition={} offset={} lagMs={}",
                    record.partition(),
                    record.offset(),
                    lagMs);
        }
        MlScoreEvent event =
                objectMapper.readValue(payload, MlScoreEvent.class);
        orchestrationService.handleMlEvent(event);
    }

    private boolean shouldLog() {
        return LOG_COUNTER.incrementAndGet() % 100 == 0;
    }
}
