package com.example.fraud_orchestrator.repository;

import java.util.Optional;

import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;

import com.example.fraud_orchestrator.entity.FraudDecisionRecord;

public interface FraudDecisionRepository extends JpaRepository<FraudDecisionRecord, Long> {
    Optional<FraudDecisionRecord> findByTransactionId(String transactionId);
    Page<FraudDecisionRecord> findByReviewedFalse(Pageable pageable);
}
