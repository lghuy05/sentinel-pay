package com.example.fraud_orchestrator.event;

import java.time.Instant;
import java.util.List;

public class FraudFinalDecisionEvent {

    private String transactionId;
    private FraudDecision decision;
    private double ruleScore;
    private double blacklistScore;
    private double mlScore;
    private double finalScore;
    private List<String> ruleMatches;
    private List<String> blacklistMatches;
    private Instant decidedAt;

    public FraudFinalDecisionEvent() {
    }

    public FraudFinalDecisionEvent(
            String transactionId,
            FraudDecision decision,
            double ruleScore,
            double blacklistScore,
            double mlScore,
            double finalScore,
            List<String> ruleMatches,
            List<String> blacklistMatches,
            Instant decidedAt
    ) {
        this.transactionId = transactionId;
        this.decision = decision;
        this.ruleScore = ruleScore;
        this.blacklistScore = blacklistScore;
        this.mlScore = mlScore;
        this.finalScore = finalScore;
        this.ruleMatches = ruleMatches;
        this.blacklistMatches = blacklistMatches;
        this.decidedAt = decidedAt;
    }

    public String getTransactionId() {
        return transactionId;
    }

    public void setTransactionId(String transactionId) {
        this.transactionId = transactionId;
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

    public Instant getDecidedAt() {
        return decidedAt;
    }

    public void setDecidedAt(Instant decidedAt) {
        this.decidedAt = decidedAt;
    }
}
