package com.example.transaction_ingestor.event.publisher;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

import com.example.transaction_ingestor.event.EventPublisher;
import com.example.transaction_ingestor.event.TransactionReceivedEvent;

@Component
@Profile("noop") // active in dev/test
public class NoOpEventPublisher implements EventPublisher {

    private static final Logger log =
            LoggerFactory.getLogger(NoOpEventPublisher.class);

    @Override
    public void publish(TransactionReceivedEvent event) {
        log.info(
            "[NOOP EVENT] transactionId={}, type={}, amount={} {}",
            event.getTransactionId(),
            event.getType(),
            event.getAmount(),
            event.getCurrency()
        );
    }
}

