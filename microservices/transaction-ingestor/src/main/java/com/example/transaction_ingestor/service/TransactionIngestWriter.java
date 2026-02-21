package com.example.transaction_ingestor.service;

import java.time.Instant;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.transaction_ingestor.dto.CreateTransactionRequest;
import com.example.transaction_ingestor.entity.OutboxEvent;
import com.example.transaction_ingestor.entity.OutboxStatusType;
import com.example.transaction_ingestor.entity.TransactionRecord;
import com.example.transaction_ingestor.event.TransactionReceivedEvent;
import com.example.transaction_ingestor.repository.OutboxEventRepository;
import com.example.transaction_ingestor.repository.TransactionIngestRepository;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

@Service
public class TransactionIngestWriter {

  private final TransactionIngestRepository transactionRepository;
  private final OutboxEventRepository outboxEventRepository;
  private final ObjectMapper objectMapper;

  public TransactionIngestWriter(
      TransactionIngestRepository transactionRepository,
      OutboxEventRepository outboxEventRepository,
      ObjectMapper objectMapper) {
    this.transactionRepository = transactionRepository;
    this.outboxEventRepository = outboxEventRepository;
    this.objectMapper = objectMapper;
  }

  @Transactional
  public TransactionRecord insertWithOutbox(CreateTransactionRequest dto) {
    Instant now = Instant.now();

    TransactionRecord record = new TransactionRecord();
    record.setTransactionId(dto.getTransactionId());
    record.setType(dto.getType());
    record.setSenderUserId(dto.getSenderUserId());
    record.setReceiverUserId(dto.getReceiverUserId());
    record.setMerchantId(dto.getMerchantId());
    record.setAmount(dto.getAmount());
    record.setCurrency(dto.getCurrency());
    record.setDeviceId(dto.getDeviceId());
    record.setEventTime(dto.getTimestamp());
    record.setReceivedAt(now);

    // Force insert/constraint evaluation in this method.
    TransactionRecord saved = transactionRepository.saveAndFlush(record);

    TransactionReceivedEvent event = new TransactionReceivedEvent(
        saved.getTransactionId(),
        saved.getType(),
        saved.getSenderUserId(),
        saved.getReceiverUserId(),
        saved.getMerchantId(),
        saved.getAmount(),
        saved.getCurrency(),
        saved.getDeviceId(),
        saved.getEventTime(),
        saved.getReceivedAt());

    OutboxEvent outbox = new OutboxEvent();
    outbox.setAggregateType("TRANSACTION");
    outbox.setAggregateId(saved.getTransactionId());
    outbox.setEventType("TransactionReceived");
    try {
      outbox.setPayload(objectMapper.writeValueAsString(event));
    } catch (JsonProcessingException e) {
      throw new IllegalStateException("Failed to serialize outbox payload", e);
    }
    outbox.setStatus(OutboxStatusType.PENDING);
    outbox.setAttemptCount(0);
    outbox.setCreatedAt(now);
    outbox.setNextRetry(now);
    outboxEventRepository.save(outbox);

    return saved;
  }
}
