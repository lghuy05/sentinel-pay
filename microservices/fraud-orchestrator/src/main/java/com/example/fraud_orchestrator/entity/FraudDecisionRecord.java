package com.example.fraud_orchestrator.entity;

import java.math.BigDecimal;
import java.time.Instant;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
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

    private Long accountId;

    @Column(precision = 19, scale = 4)
    private BigDecimal amount;

    private String country;

    @Column(columnDefinition = "text")
    private String featuresJson;

    private boolean blacklistHit;

    private Double ruleScore;

    private String ruleBand;

    @Column(columnDefinition = "text")
    private String ruleMatches;

    private Double mlScore;

    private String mlBand;

    private String finalDecision;

    private String decisionReason;

    private String modelVersion;

    private Integer ruleVersion;

    private Boolean trueLabel;

    private boolean reviewed = false;

    @Column(nullable = false)
    private Instant createdAt = Instant.now();
}
