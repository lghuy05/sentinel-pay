package com.example.fraud_orchestrator.event;

import java.time.Instant;
import java.util.List;
import java.util.Map;

import com.fasterxml.jackson.annotation.JsonAlias;
import com.fasterxml.jackson.annotation.JsonProperty;

public class RuleEvaluationEvent {

    @JsonProperty("transactionId")
    @JsonAlias("transaction_id")
    private String transactionId;
    private Long senderUserId;
    private Long receiverUserId;
    private Long merchantId;
    private java.math.BigDecimal amount;
    private String currency;
    @JsonProperty("ruleScore")
    @JsonAlias("rule_score")
    private double ruleScore;
    @JsonProperty("ruleBand")
    @JsonAlias("rule_band")
    private String ruleBand;
    @JsonProperty("ruleMatches")
    @JsonAlias("matched_rules")
    private List<String> ruleMatches;
    @JsonProperty("decisionHint")
    @JsonAlias("decision_hint")
    private String decisionHint;
    @JsonProperty("ruleVersion")
    @JsonAlias("rule_version")
    private Integer ruleVersion;
    private Map<String, Object> features;
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

    public String getRuleBand() {
        return ruleBand;
    }

    public void setRuleBand(String ruleBand) {
        this.ruleBand = ruleBand;
    }

    public List<String> getRuleMatches() {
        return ruleMatches;
    }

    public void setRuleMatches(List<String> ruleMatches) {
        this.ruleMatches = ruleMatches;
    }

    public String getDecisionHint() {
        return decisionHint;
    }

    public void setDecisionHint(String decisionHint) {
        this.decisionHint = decisionHint;
    }

    public Integer getRuleVersion() {
        return ruleVersion;
    }

    public void setRuleVersion(Integer ruleVersion) {
        this.ruleVersion = ruleVersion;
    }

    public Map<String, Object> getFeatures() {
        return features;
    }

    public void setFeatures(Map<String, Object> features) {
        this.features = features;
    }

    public Instant getEvaluatedAt() {
        return evaluatedAt;
    }

    public void setEvaluatedAt(Instant evaluatedAt) {
        this.evaluatedAt = evaluatedAt;
    }
}
