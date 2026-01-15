package com.example.transaction_ingestor.event.publisher;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import com.example.transaction_ingestor.event.EventPublisher;
import com.example.transaction_ingestor.event.TransactionReceivedEvent;

public class NoOpEventPublisher implements EventPublisher {

    private static final Logger log =
            LoggerFactory.getLogger(NoOpEventPublisher.class);

    @Override
    public void publish(TransactionReceivedEvent event) {
        log.info("NOOP publisher received event: {}", event);
    }
}
