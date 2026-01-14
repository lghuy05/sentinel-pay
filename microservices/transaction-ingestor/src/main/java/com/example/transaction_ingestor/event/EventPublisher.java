package com.example.transaction_ingestor.event;

public interface EventPublisher {
    void publish(TransactionReceivedEvent event);
}
