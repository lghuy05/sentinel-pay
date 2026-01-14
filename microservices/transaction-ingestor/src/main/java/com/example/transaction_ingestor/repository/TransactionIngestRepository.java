package com.example.transaction_ingestor.repository;

import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;

import com.example.transaction_ingestor.entity.TransactionRecord;

public interface TransactionIngestRepository
        extends JpaRepository<TransactionRecord, Long> {

    Optional<TransactionRecord> findByTransactionId(String transactionId);
}

