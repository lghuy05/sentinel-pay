package com.example.fraud_orchestrator.event;

import java.time.Instant;
import java.util.List;

public class FraudFinalDecisionEvent {

    private String transactionId;
    private Long senderUserId;
    private Long receiverUserId;
    private Long merchantId;
    private java.math.BigDecimal amount;
    private String currency;
    private FraudDecision decision;
    private double ruleScore;
    private double blacklistScore;
    private double mlScore;
    private double finalScore;
    private List<String> ruleMatches;
    private List<String> blacklistMatches;
    private int riskScore;
    private String riskLevel;
    private List<String> triggeredRules;
    private List<String> hardStopMatches;
    private String hardStopDecision;
    private Instant decidedAt;

    public FraudFinalDecisionEvent() {
    }

    public FraudFinalDecisionEvent(
            String transactionId,
            Long senderUserId,
            Long receiverUserId,
            Long merchantId,
            java.math.BigDecimal amount,
            String currency,
            FraudDecision decision,
            double ruleScore,
            double blacklistScore,
            double mlScore,
            double finalScore,
            List<String> ruleMatches,
            List<String> blacklistMatches,
            int riskScore,
            String riskLevel,
            List<String> triggeredRules,
            List<String> hardStopMatches,
            String hardStopDecision,
            Instant decidedAt
    ) {
        this.transactionId = transactionId;
        this.senderUserId = senderUserId;
        this.receiverUserId = receiverUserId;
        this.merchantId = merchantId;
        this.amount = amount;
        this.currency = currency;
        this.decision = decision;
        this.ruleScore = ruleScore;
        this.blacklistScore = blacklistScore;
        this.mlScore = mlScore;
        this.finalScore = finalScore;
        this.ruleMatches = ruleMatches;
        this.blacklistMatches = blacklistMatches;
        this.riskScore = riskScore;
        this.riskLevel = riskLevel;
        this.triggeredRules = triggeredRules;
        this.hardStopMatches = hardStopMatches;
        this.hardStopDecision = hardStopDecision;
        this.decidedAt = decidedAt;
    }

    public String getTransactionId() {
        return transactionId;
    }

    public void setTransactionId(String transactionId) {
        this.transactionId = transactionId;
    }

    public Long getSenderUserId() {
        return senderUserId;
    }

    public void setSenderUserId(Long senderUserId) {
        this.senderUserId = senderUserId;
    }

    public Long getReceiverUserId() {
        return receiverUserId;
    }

    public void setReceiverUserId(Long receiverUserId) {
        this.receiverUserId = receiverUserId;
    }

    public Long getMerchantId() {
        return merchantId;
    }

    public void setMerchantId(Long merchantId) {
        this.merchantId = merchantId;
    }

    public java.math.BigDecimal getAmount() {
        return amount;
    }

    public void setAmount(java.math.BigDecimal amount) {
        this.amount = amount;
    }

    public String getCurrency() {
        return currency;
    }

    public void setCurrency(String currency) {
        this.currency = currency;
    }

    public FraudDecision getDecision() {
        return decision;
    }

    public void setDecision(FraudDecision decision) {
        this.decision = decision;
    }

    public double getRuleScore() {
        return ruleScore;
    }

    public void setRuleScore(double ruleScore) {
        this.ruleScore = ruleScore;
    }

    public double getBlacklistScore() {
        return blacklistScore;
    }

    public void setBlacklistScore(double blacklistScore) {
        this.blacklistScore = blacklistScore;
    }

    public double getMlScore() {
        return mlScore;
    }

    public void setMlScore(double mlScore) {
        this.mlScore = mlScore;
    }

    public double getFinalScore() {
        return finalScore;
    }

    public void setFinalScore(double finalScore) {
        this.finalScore = finalScore;
    }

    public List<String> getRuleMatches() {
        return ruleMatches;
    }

    public void setRuleMatches(List<String> ruleMatches) {
        this.ruleMatches = ruleMatches;
    }

    public List<String> getBlacklistMatches() {
        return blacklistMatches;
    }

    public void setBlacklistMatches(List<String> blacklistMatches) {
        this.blacklistMatches = blacklistMatches;
    }

    public int getRiskScore() {
        return riskScore;
    }

    public void setRiskScore(int riskScore) {
        this.riskScore = riskScore;
    }

    public String getRiskLevel() {
        return riskLevel;
    }

    public void setRiskLevel(String riskLevel) {
        this.riskLevel = riskLevel;
    }

    public List<String> getTriggeredRules() {
        return triggeredRules;
    }

    public void setTriggeredRules(List<String> triggeredRules) {
        this.triggeredRules = triggeredRules;
    }

    public List<String> getHardStopMatches() {
        return hardStopMatches;
    }

    public void setHardStopMatches(List<String> hardStopMatches) {
        this.hardStopMatches = hardStopMatches;
    }

    public String getHardStopDecision() {
        return hardStopDecision;
    }

    public void setHardStopDecision(String hardStopDecision) {
        this.hardStopDecision = hardStopDecision;
    }

    public Instant getDecidedAt() {
        return decidedAt;
    }

    public void setDecidedAt(Instant decidedAt) {
        this.decidedAt = decidedAt;
    }
}
