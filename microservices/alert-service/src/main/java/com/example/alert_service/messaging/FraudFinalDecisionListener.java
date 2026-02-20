package com.example.alert_service.messaging;

import com.example.alert_service.event.FraudFinalDecisionEvent;
import com.example.alert_service.service.FraudDecisionService;
import com.fasterxml.jackson.databind.ObjectMapper;

import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

import java.util.concurrent.atomic.AtomicLong;

@Component
public class FraudFinalDecisionListener {

    private static final Logger log = LoggerFactory.getLogger(FraudFinalDecisionListener.class);
    private static final AtomicLong LOG_COUNTER = new AtomicLong(0);

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
    public void handleFinalDecision(ConsumerRecord<String, String> record) throws Exception {
        String payload = record.value();
        if (shouldLog()) {
            long lagMs = record.timestamp() > 0 ? System.currentTimeMillis() - record.timestamp() : -1;
            log.info("Kafka consume topic=fraud.final partition={} offset={} lagMs={}",
                    record.partition(),
                    record.offset(),
                    lagMs);
        }
        FraudFinalDecisionEvent event = objectMapper.readValue(payload, FraudFinalDecisionEvent.class);
        fraudDecisionService.handleDecision(event);
    }

    private boolean shouldLog() {
        return LOG_COUNTER.incrementAndGet() % 100 == 0;
    }
}
