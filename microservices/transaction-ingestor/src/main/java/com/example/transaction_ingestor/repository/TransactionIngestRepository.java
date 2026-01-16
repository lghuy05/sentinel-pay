package com.example.transaction_ingestor.repository;

import java.util.Optional;
import java.util.List;
import java.time.Instant;

import org.springframework.data.jpa.repository.JpaRepository;

import com.example.transaction_ingestor.entity.TransactionRecord;

public interface TransactionIngestRepository
        extends JpaRepository<TransactionRecord, Long> {

    Optional<TransactionRecord> findByTransactionId(String transactionId);

    List<TransactionRecord> findTop50ByOrderByReceivedAtDesc();

    long countBySenderUserIdAndEventTimeBetweenAndAmountGreaterThanEqualAndCurrencyIgnoreCase(
            Long senderUserId,
            Instant start,
            Instant end,
            java.math.BigDecimal amount,
            String currency
    );
}
