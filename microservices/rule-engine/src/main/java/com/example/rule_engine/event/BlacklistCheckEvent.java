package com.example.rule_engine.event;

import java.time.Instant;

public class BlacklistCheckEvent {

    private String transactionId;
    private boolean blacklistHit;
    private String reason;
    private String decisionHint;
    private TransactionEnrichedEvent transaction;
    private Instant evaluatedAt;

    public BlacklistCheckEvent() {
    }

    public String getTransactionId() {
        return transactionId;
    }

    public void setTransactionId(String transactionId) {
        this.transactionId = transactionId;
    }

    public boolean isBlacklistHit() {
        return blacklistHit;
    }

    public void setBlacklistHit(boolean blacklistHit) {
        this.blacklistHit = blacklistHit;
    }

    public String getReason() {
        return reason;
    }

    public void setReason(String reason) {
        this.reason = reason;
    }

    public String getDecisionHint() {
        return decisionHint;
    }

    public void setDecisionHint(String decisionHint) {
        this.decisionHint = decisionHint;
    }

    public TransactionEnrichedEvent getTransaction() {
        return transaction;
    }

    public void setTransaction(TransactionEnrichedEvent transaction) {
        this.transaction = transaction;
    }

    public Instant getEvaluatedAt() {
        return evaluatedAt;
    }

    public void setEvaluatedAt(Instant evaluatedAt) {
        this.evaluatedAt = evaluatedAt;
    }
}
