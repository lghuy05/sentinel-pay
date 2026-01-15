package com.example.transaction_ingestor.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;

import com.example.transaction_ingestor.event.EventPublisher;
import com.example.transaction_ingestor.event.TransactionReceivedEvent;
import com.example.transaction_ingestor.event.publisher.KafkaEventPublisher;
import com.example.transaction_ingestor.event.publisher.NoOpEventPublisher;


@Configuration
public class EventPublisherConfig {

    @Bean
    @ConditionalOnProperty(
        name = "event.publisher",
        havingValue = "kafka"
    )
    public EventPublisher kafkaPublisher(
            KafkaTemplate<String, TransactionReceivedEvent> kafkaTemplate
    ) {
        return new KafkaEventPublisher(kafkaTemplate);
    }

    @Bean
    @ConditionalOnMissingBean(EventPublisher.class)
    public EventPublisher noopPublisher() {
        return new NoOpEventPublisher();
    }
}

