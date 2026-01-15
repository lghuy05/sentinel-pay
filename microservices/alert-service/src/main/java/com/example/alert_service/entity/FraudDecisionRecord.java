package com.example.alert_service.entity;

import java.time.Instant;

import com.example.alert_service.event.FraudDecision;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;

import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
@Entity
@Table(name = "fraud_decisions")
public class FraudDecisionRecord {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, unique = true)
    private String transactionId;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private FraudDecision decision;

    @Column(nullable = false)
    private double ruleScore;

    @Column(nullable = false)
    private double blacklistScore;

    @Column(nullable = false)
    private double mlScore;

    @Column(nullable = false)
    private double finalScore;

    @Column(columnDefinition = "text")
    private String ruleMatches;

    @Column(columnDefinition = "text")
    private String blacklistMatches;

    @Column(nullable = false)
    private Instant decidedAt;

    @Column(nullable = false)
    private Instant createdAt = Instant.now();
}
