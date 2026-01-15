package com.example.transaction_ingestor.event.publisher;

import com.example.transaction_ingestor.event.EventPublisher;
import com.example.transaction_ingestor.event.TransactionReceivedEvent;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.kafka.core.KafkaTemplate;
public class KafkaEventPublisher implements EventPublisher {

    private static final Logger log =
            LoggerFactory.getLogger(KafkaEventPublisher.class);

    private static final String TOPIC = "transactions.raw";

    private final KafkaTemplate<String, TransactionReceivedEvent> kafkaTemplate;

    public KafkaEventPublisher(
            KafkaTemplate<String, TransactionReceivedEvent> kafkaTemplate
    ) {
        this.kafkaTemplate = kafkaTemplate;
    }

    public void publish(TransactionReceivedEvent event) {
        try {
            kafkaTemplate.send(
                    TOPIC,
                    event.getTransactionId(), // key
                    event
            );

            log.info(
                    "Kafka published TransactionReceivedEvent txId={}",
                    event.getTransactionId()
            );
        } catch (Exception e) {
            log.error(
                    "Kafka publish FAILED txId={}",
                    event.getTransactionId(),
                    e
            );
        }
    }
}
