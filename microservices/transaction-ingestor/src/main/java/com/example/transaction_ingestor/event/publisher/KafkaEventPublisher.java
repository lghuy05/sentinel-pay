package com.example.transaction_ingestor.event.publisher;

import com.example.transaction_ingestor.event.EventPublisher;
import com.example.transaction_ingestor.event.TransactionReceivedEvent;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.atomic.AtomicLong;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.kafka.support.SendResult;
public class KafkaEventPublisher implements EventPublisher {

    private static final Logger log =
            LoggerFactory.getLogger(KafkaEventPublisher.class);

    private static final String TOPIC = "transactions.raw";
    private static final AtomicLong LOG_COUNTER = new AtomicLong(0);

    private final KafkaTemplate<String, TransactionReceivedEvent> kafkaTemplate;

    public KafkaEventPublisher(
            KafkaTemplate<String, TransactionReceivedEvent> kafkaTemplate
    ) {
        this.kafkaTemplate = kafkaTemplate;
    }

    public void publish(TransactionReceivedEvent event) {
        try {
            long startNs = System.nanoTime();
            CompletableFuture<SendResult<String, TransactionReceivedEvent>> future = kafkaTemplate.send(
                    TOPIC,
                    event.getTransactionId(), // key
                    event
            );
            future.whenComplete((result, ex) -> {
                if (ex != null) {
                    log.error("Kafka publish FAILED txId={}", event.getTransactionId(), ex);
                    return;
                }
                if (shouldLog()) {
                    long ackMs = (System.nanoTime() - startNs) / 1_000_000;
                    log.info(
                            "Kafka published TransactionReceivedEvent txId={} partition={} offset={} ackMs={}",
                            event.getTransactionId(),
                            result.getRecordMetadata().partition(),
                            result.getRecordMetadata().offset(),
                            ackMs
                    );
                }
            });
        } catch (Exception e) {
            log.error(
                    "Kafka publish FAILED txId={}",
                    event.getTransactionId(),
                    e
            );
        }
    }

    private boolean shouldLog() {
        return LOG_COUNTER.incrementAndGet() % 100 == 0;
    }
}
