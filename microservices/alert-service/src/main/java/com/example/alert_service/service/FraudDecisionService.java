package com.example.alert_service.service;

import java.time.Instant;
import java.time.chrono.IsoChronology;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.alert_service.entity.FraudDecisionRecord;
import com.example.alert_service.entity.Transfer;
import com.example.alert_service.entity.TransferStatus;
import com.example.alert_service.event.FraudDecision;
import com.example.alert_service.event.FraudFinalDecisionEvent;
import com.example.alert_service.repository.FraudDecisionRepository;
import com.example.alert_service.repository.TransferRepository;
import com.fasterxml.jackson.databind.ObjectMapper;

@Service
public class FraudDecisionService {

  private final FraudDecisionRepository repository;
  private final TransferRepository transferRepository;
  private final NotificationService notificationService;
  private final AccountBalanceService accountBalanceService;
  private final ObjectMapper objectMapper;

  public FraudDecisionService(
      FraudDecisionRepository repository,
      NotificationService notificationService,
      AccountBalanceService accountBalanceService,
      TransferRepository transferRepository,
      ObjectMapper objectMapper) {
    this.repository = repository;
    this.notificationService = notificationService;
    this.accountBalanceService = accountBalanceService;
    this.transferRepository = transferRepository;
    this.objectMapper = objectMapper;
  }

  @Transactional
  public void handleDecision(FraudFinalDecisionEvent event) {
    if (event == null || event.getFinalDecision() == null) {
      return;
    }

    if (event.getFinalDecision() == FraudDecision.ALLOW) {
      handleAllowedTransfer(event);
      return;
    }

    FraudDecisionRecord record = new FraudDecisionRecord();
    record.setTransactionId(event.getTransactionId());
    record.setDecision(event.getFinalDecision());
    record.setDecisionReason(event.getDecisionReason());
    record.setPayloadJson(serialize(event));
    record.setDecidedAt(event.getDecidedAt() != null ? event.getDecidedAt() : Instant.now());

    repository.save(record);
    notificationService.sendAlert(event);
  }

  private void handleAllowedTransfer(FraudFinalDecisionEvent event) {
    String transactionId = event.getTransactionId();
    if (transactionId == null || transactionId.isBlank()) {
      return;
    }

    Transfer transfer = transferRepository
        .findByTransacionIdForUpdate(transactionId)
        .orElseGet(() -> {
          Transfer t = new Transfer();
          t.setTransactionId(transactionId);
          t.setStatus(TransferStatus.PROCESSING);
          t.setAttempts(0);
          t.setCreatedAt(Instant.now());
          t.setUpdatedAt(Instant.now());
          return t;
        });

    if (transfer.getStatus() == TransferStatus.APPLIED) {
      return;
    }

    transfer.setStatus(TransferStatus.PROCESSING);
    transfer.setAttempts(transfer.getAttempts() + 1);
    transfer.setLastError(null);
    transfer.setUpdatedAt(Instant.now());
    transferRepository.save(transfer);

    try {
      accountBalanceService.applyIfAllowed(event);
      transfer.setStatus(TransferStatus.APPLIED);
      transfer.setUpdatedAt(Instant.now());
      transferRepository.save(transfer);
    } catch (Exception ex) {
      transfer.setStatus(TransferStatus.FAILED_RETRYABLE);
      transfer.setLastError(truncate(ex.getMessage()));
      transfer.setUpdatedAt(Instant.now());
      transferRepository.save(transfer);
      throw ex;
    }
  }

  private String truncate(String message) {
    if (message == null) {
      return null;
    }
    return message.length() > 500 ? message.substring(0, 500) : message;
  }

  private String serialize(Object value) {
    try {
      return objectMapper.writeValueAsString(value);
    } catch (Exception e) {
      return "[]";
    }
  }
}
