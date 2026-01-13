package com.example.transaction_ingestor.service;

import java.lang.Override;
import org.springframework.beans.factory.annotation.Autowired;
import com.example.transaction_ingestor.entity.TransactionRecord;

public interface TransactionIngestorService {
    TransactionRecord createTransaction(TransactionRecord transaction);
}
