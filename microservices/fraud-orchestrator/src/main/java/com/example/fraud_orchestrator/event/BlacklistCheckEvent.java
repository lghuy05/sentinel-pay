package com.example.fraud_orchestrator.event;

import java.time.Instant;
import java.util.List;

public class BlacklistCheckEvent {

    private String transactionId;
    private double blacklistScore;
    private List<String> matchedEntries;
    private Instant evaluatedAt;

    public BlacklistCheckEvent() {
    }

    public String getTransactionId() {
        return transactionId;
    }

    public void setTransactionId(String transactionId) {
        this.transactionId = transactionId;
    }

    public double getBlacklistScore() {
        return blacklistScore;
    }

    public void setBlacklistScore(double blacklistScore) {
        this.blacklistScore = blacklistScore;
    }

    public List<String> getMatchedEntries() {
        return matchedEntries;
    }

    public void setMatchedEntries(List<String> matchedEntries) {
        this.matchedEntries = matchedEntries;
    }

    public Instant getEvaluatedAt() {
        return evaluatedAt;
    }

    public void setEvaluatedAt(Instant evaluatedAt) {
        this.evaluatedAt = evaluatedAt;
    }
}
