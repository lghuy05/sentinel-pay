package com.example.alert_service.event;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.List;

public class FraudFinalDecisionEvent {

    private String transactionId;
    private Long accountId;
    private Long senderUserId;
    private Long receiverUserId;
    private Long merchantId;
    private BigDecimal amount;
    private String currency;
    private String country;
    private String featuresJson;
    private boolean blacklistHit;
    private Double ruleScore;
    private String ruleBand;
    private List<String> ruleMatches;
    private Double mlScore;
    private String mlBand;
    private FraudDecision finalDecision;
    private String decisionReason;
    private String modelVersion;
    private Integer ruleVersion;
    private Instant decidedAt;

    public FraudFinalDecisionEvent() {
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

    public BigDecimal getAmount() {
        return amount;
    }

    public void setAmount(BigDecimal amount) {
        this.amount = amount;
    }

    public String getCurrency() {
        return currency;
    }

    public void setCurrency(String currency) {
        this.currency = currency;
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

    public List<String> getRuleMatches() {
        return ruleMatches;
    }

    public void setRuleMatches(List<String> ruleMatches) {
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

    public FraudDecision getFinalDecision() {
        return finalDecision;
    }

    public void setFinalDecision(FraudDecision finalDecision) {
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

    public Instant getDecidedAt() {
        return decidedAt;
    }

    public void setDecidedAt(Instant decidedAt) {
        this.decidedAt = decidedAt;
    }
}
