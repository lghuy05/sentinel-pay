package com.example.transaction_ingestor.service;

import java.time.Instant;
import java.util.List;

import org.springframework.data.domain.PageRequest;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.transaction_ingestor.entity.OutboxEvent;
import com.example.transaction_ingestor.entity.OutboxStatusType;
import com.example.transaction_ingestor.event.TransactionReceivedEvent;
import com.example.transaction_ingestor.repository.OutboxEventRepository;
import com.fasterxml.jackson.databind.ObjectMapper;

@Service
public class OutboxRelayService {

  private static final int MAX_ATTEMPTS = 5;
  private static final int BASE_BACKOFF_SECONDS = 5;

  private final OutboxEventRepository outboxEventRepository;
  private final KafkaTemplate<String, TransactionReceivedEvent> kafkaTemplate;
  private final ObjectMapper objectMapper;

  public OutboxRelayService(
      OutboxEventRepository outboxEventRepository,
      KafkaTemplate<String, TransactionReceivedEvent> kafkaTemplate,
      ObjectMapper objectMapper) {
    this.outboxEventRepository = outboxEventRepository;
    this.kafkaTemplate = kafkaTemplate;
    this.objectMapper = objectMapper;
  }

  @Scheduled(fixedDelayString = "${outbox.relay.delay-ms:1000}")
  @Transactional
  public void relay() {
    List<OutboxEvent> batch = outboxEventRepository.findDue(
        OutboxStatusType.PENDING,
        Instant.now(),
        PageRequest.of(0, 100));
    for (OutboxEvent o : batch) {

      try {
        TransactionReceivedEvent evt = objectMapper.readValue(o.getPayload(), TransactionReceivedEvent.class);
        kafkaTemplate.send("transactions.raw",
            o.getAggregateId(), evt).get();
        o.setStatus(OutboxStatusType.SENT);
        o.setPublishedAt(Instant.now());
        o.setLastError(null);
        o.setNextRetry(null);
      } catch (Exception ex) {
        int attempts = o.getAttemptCount() + 1;
        o.setAttemptCount(attempts);
        o.setLastError(truncate(ex.getMessage()));
        if (attempts >= MAX_ATTEMPTS) {
          o.setStatus(OutboxStatusType.FAILED);
        } else {
          o.setStatus(OutboxStatusType.PENDING);
          o.setNextRetry(Instant.now().plusSeconds(backoffSeconds(attempts)));
        }
      }
      outboxEventRepository.save(o);
    }
  }

  private long backoffSeconds(int attempts) {
    return (long) BASE_BACKOFF_SECONDS * attempts;
  }

  private String truncate(String message) {
    if (message == null) {
      return null;
    }
    return message.length() > 500 ? message.substring(0, 500) : message;
  }
}
