package com.example.transaction_ingestor.repository.impl;

import com.example.transaction_ingestor.repository.TransactionIngestorRepository;
import org.springframework.beans.factory.annotation.Autowired;
import java.lang.Override;

public class TransactionIngestorRepositoryImpl {

    @Autowired
    private TransactionIngestorRepository transactionIngestorRepository;

    @Override
    public TransactionRecord createTransaction(TransactionRecord transaction){
        return 
    }

}
