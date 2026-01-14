package com.example.transaction_ingestor.service;

import java.time.Instant;
import java.util.Optional;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.transaction_ingestor.dto.CreateTransactionRequest;
import com.example.transaction_ingestor.dto.TransactionType;
import com.example.transaction_ingestor.entity.TransactionRecord;
import com.example.transaction_ingestor.event.EventPublisher;
import com.example.transaction_ingestor.event.TransactionReceivedEvent;
import com.example.transaction_ingestor.repository.TransactionIngestRepository;
import com.example.transaction_ingestor.repository.TransactionIngestRepository;

@Service
public class TransactionIngestService {

    private final TransactionIngestRepository transactionRepository;
    private final EventPublisher eventPublisher;

    public TransactionIngestService(
            TransactionIngestRepository transactionRepository,
            EventPublisher eventPublisher
    ) {
        this.transactionRepository = transactionRepository;
        this.eventPublisher = eventPublisher;
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
    @Transactional
    public TransactionRecord ingest(CreateTransactionRequest dto) {

        // 1️⃣ Business validation (input correctness, NOT fraud logic)
        validateBusinessRules(dto);

        // 2️⃣ Idempotency check
        Optional<TransactionRecord> existing =
                transactionRepository.findByTransactionId(dto.getTransactionId());

        if (existing.isPresent()) {
            // Safe retry — return already-ingested record
            return existing.get();
        }

        // 3️⃣ Map DTO → Entity (fact creation)
        TransactionRecord record = new TransactionRecord();
        record.setTransactionId(dto.getTransactionId());
        record.setType(dto.getType());
        record.setSenderUserId(dto.getSenderUserId());
        record.setReceiverUserId(dto.getReceiverUserId());
        record.setMerchantId(dto.getMerchantId());
        record.setAmount(dto.getAmount());
        record.setCurrency(dto.getCurrency());
        record.setIp(dto.getIp());
        record.setDeviceId(dto.getDeviceId());
        record.setEventTime(dto.getTimestamp());
        record.setReceivedAt(Instant.now());

        // 4️⃣ Persist FIRST (DB = source of truth)
        TransactionRecord saved = transactionRepository.save(record);

        // 5️⃣ Build domain event from persisted data
        TransactionReceivedEvent event =
                new TransactionReceivedEvent(
                        saved.getTransactionId(),
                        saved.getType(),
                        saved.getSenderUserId(),
                        saved.getReceiverUserId(),
                        saved.getMerchantId(),
                        saved.getAmount(),
                        saved.getCurrency(),
                        saved.getIp(),
                        saved.getDeviceId(),
                        saved.getEventTime(),
                        saved.getReceivedAt()
                );

        // 6️⃣ Emit event (NoOp now, Kafka later)
        eventPublisher.publish(event);

        return saved;
    }

    /**
     * Business-level validation.
     * This ensures the input makes sense structurally.
     * This is NOT fraud detection.
     */
    private void validateBusinessRules(CreateTransactionRequest dto) {

        if (dto.getType() == TransactionType.MERCHANT_PAYMENT) {
            if (dto.getMerchantId() == null) {
                throw new IllegalArgumentException(
                        "merchantId is required for MERCHANT_PAYMENT"
                );
            }
            if (dto.getReceiverUserId() != null) {
                throw new IllegalArgumentException(
                        "receiverUserId must be null for MERCHANT_PAYMENT"
                );
            }
        }

        if (dto.getType() == TransactionType.P2P_TRANSFER) {
            if (dto.getReceiverUserId() == null) {
                throw new IllegalArgumentException(
                        "receiverUserId is required for P2P_TRANSFER"
                );
            }
            if (dto.getMerchantId() != null) {
                throw new IllegalArgumentException(
                        "merchantId must be null for P2P_TRANSFER"
                );
            }
        }
    }
}

