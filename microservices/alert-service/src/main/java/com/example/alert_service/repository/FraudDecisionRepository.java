package com.example.alert_service.repository;

import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;

import com.example.alert_service.entity.FraudDecisionRecord;

public interface FraudDecisionRepository extends JpaRepository<FraudDecisionRecord, Long> {
    Optional<FraudDecisionRecord> findByTransactionId(String transactionId);
}
