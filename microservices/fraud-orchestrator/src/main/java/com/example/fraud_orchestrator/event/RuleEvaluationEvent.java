package com.example.fraud_orchestrator.event;

import java.time.Instant;
import java.util.List;

public class RuleEvaluationEvent {

    private String transactionId;
    private Long senderUserId;
    private Long receiverUserId;
    private Long merchantId;
    private java.math.BigDecimal amount;
    private String currency;
    private double ruleScore;
    private List<String> matchedRules;
    private int riskScore;
    private String riskLevel;
    private List<String> triggeredRules;
    private List<String> hardStopMatches;
    private String hardStopDecision;
    private Instant evaluatedAt;

    public RuleEvaluationEvent() {
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

    public double getRuleScore() {
        return ruleScore;
    }

    public void setRuleScore(double ruleScore) {
        this.ruleScore = ruleScore;
    }

    public List<String> getMatchedRules() {
        return matchedRules;
    }

    public void setMatchedRules(List<String> matchedRules) {
        this.matchedRules = matchedRules;
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

    public Instant getEvaluatedAt() {
        return evaluatedAt;
    }

    public void setEvaluatedAt(Instant evaluatedAt) {
        this.evaluatedAt = evaluatedAt;
    }
}
