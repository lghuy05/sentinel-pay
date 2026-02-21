package com.example.transaction_ingestor.service;

import java.time.Instant;
import java.time.LocalDate;
import java.time.ZoneOffset;
import java.util.Optional;
import java.math.BigDecimal;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.transaction_ingestor.dto.CreateTransactionRequest;
import com.example.transaction_ingestor.dto.TransactionType;
import com.example.transaction_ingestor.entity.TransactionRecord;
import com.example.transaction_ingestor.repository.TransactionIngestRepository;

@Service
public class TransactionIngestService {

  private static final BigDecimal HIGH_VALUE_USD_THRESHOLD = new BigDecimal("500");
  private static final int DAILY_HIGH_VALUE_LIMIT = 2;

  private final TransactionIngestRepository transactionRepository;
  private final TransactionIngestWriter transactionIngestWriter;
  private final boolean highValueLimitEnabled;

  public TransactionIngestService(
      TransactionIngestRepository transactionRepository,
      TransactionIngestWriter transactionIngestWriter,
      @Value("${ingest.highValueLimit.enabled:true}") boolean highValueLimitEnabled) {
    this.transactionRepository = transactionRepository;
    this.transactionIngestWriter = transactionIngestWriter;
    this.highValueLimitEnabled = highValueLimitEnabled;
  }

  /**
   * Ingest a transaction into SentinelPay.
   *
   * Responsibilities:
   * 1. Business validation (P2P vs Merchant)
   * 2. Idempotency check
   * 3. Enrichment (receivedAt)
   * 4. Persistence
   * 5. Event creation
   * 6. Event emission
   */

  public TransactionRecord ingest(CreateTransactionRequest dto) {

    // 1️⃣ Business validation
    validateBusinessRules(dto);

    // 2️⃣ Idempotency check
    Optional<TransactionRecord> existing = transactionRepository.findByTransactionId(dto.getTransactionId());

    if (existing.isPresent()) {
      // Safe retry — return already-ingested record
      return existing.get();
    }

    try {
      return transactionIngestWriter.insertWithOutbox(dto);
    } catch (DataIntegrityViolationException ex) {
      return transactionRepository.findByTransactionId(dto.getTransactionId())
          .orElseThrow(() -> new IllegalStateException(
              "Duplicate transactionId detected but existing record not found",
              ex));
    }
  }

  @Transactional(readOnly = true)
  public java.util.List<TransactionRecord> listRecent(int limit) {
    int capped = Math.max(1, Math.min(limit, 200));
    if (capped <= 50) {
      java.util.List<TransactionRecord> records = transactionRepository.findTop50ByOrderByReceivedAtDesc();
      return records.subList(0, Math.min(capped, records.size()));
    }
    return transactionRepository.findAll(
        org.springframework.data.domain.PageRequest.of(0, capped,
            org.springframework.data.domain.Sort.by("receivedAt").descending()))
        .getContent();
  }

  /**
   * Business-level validation.
   * This ensures the input makes sense structurally.
   * This is NOT fraud detection.
   */
  private void validateBusinessRules(CreateTransactionRequest dto) {

    if (dto.getAmount() == null || dto.getAmount().compareTo(BigDecimal.ZERO) <= 0) {
      throw new IllegalArgumentException("amount must be positive");
    }

    if (dto.getType() == TransactionType.MERCHANT_PAYMENT) {
      if (dto.getMerchantId() == null) {
        throw new IllegalArgumentException(
            "merchantId is required for MERCHANT_PAYMENT");
      }
      if (dto.getReceiverUserId() != null) {
        throw new IllegalArgumentException(
            "receiverUserId must be null for MERCHANT_PAYMENT");
      }
    }

    if (dto.getType() == TransactionType.P2P_TRANSFER) {
      if (dto.getReceiverUserId() == null) {
        throw new IllegalArgumentException(
            "receiverUserId is required for P2P_TRANSFER");
      }
      if (dto.getMerchantId() != null) {
        throw new IllegalArgumentException(
            "merchantId must be null for P2P_TRANSFER");
      }
    }

    enforceDailyHighValueLimit(dto);
  }

  private void enforceDailyHighValueLimit(CreateTransactionRequest dto) {
    if (!highValueLimitEnabled) {
      return;
    }
    if (dto.getSenderUserId() == null || dto.getTimestamp() == null) {
      return;
    }
    if (dto.getAmount() == null || dto.getCurrency() == null) {
      return;
    }
    if (!"USD".equalsIgnoreCase(dto.getCurrency())) {
      return;
    }
    if (dto.getAmount().compareTo(HIGH_VALUE_USD_THRESHOLD) <= 0) {
      return;
    }

    LocalDate day = dto.getTimestamp().atZone(ZoneOffset.UTC).toLocalDate();
    Instant start = day.atStartOfDay(ZoneOffset.UTC).toInstant();
    Instant end = day.plusDays(1).atStartOfDay(ZoneOffset.UTC).toInstant();

    long count = transactionRepository
        .countBySenderUserIdAndEventTimeBetweenAndAmountGreaterThanEqualAndCurrencyIgnoreCase(
            dto.getSenderUserId(),
            start,
            end,
            HIGH_VALUE_USD_THRESHOLD,
            "USD");

    if (count >= DAILY_HIGH_VALUE_LIMIT) {
      throw new IllegalArgumentException(
          "Daily limit exceeded for high-value USD transactions");
    }
  }
}
