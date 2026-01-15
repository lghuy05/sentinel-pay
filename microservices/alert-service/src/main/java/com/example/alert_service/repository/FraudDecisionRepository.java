package com.example.alert_service.repository;

import org.springframework.data.jpa.repository.JpaRepository;

import com.example.alert_service.entity.FraudDecisionRecord;

public interface FraudDecisionRepository extends JpaRepository<FraudDecisionRecord, Long> {
}
