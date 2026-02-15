package com.example.fraud_orchestrator.entity;

import java.math.BigDecimal;
import java.time.Instant;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;

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

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getTransactionId() {
        return transactionId;
    }

    public void setTransactionId(String transactionId) {
        this.transactionId = transactionId;
    }

    public Long getAccountId() {
        return accountId;
    }

    public void setAccountId(Long accountId) {
        this.accountId = accountId;
    }

    public BigDecimal getAmount() {
        return amount;
    }

    public void setAmount(BigDecimal amount) {
        this.amount = amount;
    }

    public String getCountry() {
        return country;
    }

    public void setCountry(String country) {
        this.country = country;
    }

    public String getFeaturesJson() {
        return featuresJson;
    }

    public void setFeaturesJson(String featuresJson) {
        this.featuresJson = featuresJson;
    }

    public boolean isBlacklistHit() {
        return blacklistHit;
    }

    public void setBlacklistHit(boolean blacklistHit) {
        this.blacklistHit = blacklistHit;
    }

    public Double getRuleScore() {
        return ruleScore;
    }

    public void setRuleScore(Double ruleScore) {
        this.ruleScore = ruleScore;
    }

    public String getRuleBand() {
        return ruleBand;
    }

    public void setRuleBand(String ruleBand) {
        this.ruleBand = ruleBand;
    }

    public String getRuleMatches() {
        return ruleMatches;
    }

    public void setRuleMatches(String ruleMatches) {
        this.ruleMatches = ruleMatches;
    }

    public Double getMlScore() {
        return mlScore;
    }

    public void setMlScore(Double mlScore) {
        this.mlScore = mlScore;
    }

    public String getMlBand() {
        return mlBand;
    }

    public void setMlBand(String mlBand) {
        this.mlBand = mlBand;
    }

    public String getFinalDecision() {
        return finalDecision;
    }

    public void setFinalDecision(String finalDecision) {
        this.finalDecision = finalDecision;
    }

    public String getDecisionReason() {
        return decisionReason;
    }

    public void setDecisionReason(String decisionReason) {
        this.decisionReason = decisionReason;
    }

    public String getModelVersion() {
        return modelVersion;
    }

    public void setModelVersion(String modelVersion) {
        this.modelVersion = modelVersion;
    }

    public Integer getRuleVersion() {
        return ruleVersion;
    }

    public void setRuleVersion(Integer ruleVersion) {
        this.ruleVersion = ruleVersion;
    }

    public Boolean getTrueLabel() {
        return trueLabel;
    }

    public void setTrueLabel(Boolean trueLabel) {
        this.trueLabel = trueLabel;
    }

    public boolean isReviewed() {
        return reviewed;
    }

    public void setReviewed(boolean reviewed) {
        this.reviewed = reviewed;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Instant createdAt) {
        this.createdAt = createdAt;
    }
}
