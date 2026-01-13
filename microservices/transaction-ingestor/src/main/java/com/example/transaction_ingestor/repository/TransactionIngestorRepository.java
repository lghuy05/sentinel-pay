package com.example.transaction_ingestor.repository;

import com.example.transaction_ingestor.entity.TransactionRecord;

public interface TransactionIngestorRepository {
    TransactionRecord createTransaction(TransactionRecord transaction);
}

