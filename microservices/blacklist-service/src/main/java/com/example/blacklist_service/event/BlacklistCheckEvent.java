package com.example.blacklist_service.event;

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

    public BlacklistCheckEvent(
            String transactionId,
            boolean blacklistHit,
            String reason,
            String decisionHint,
            TransactionEnrichedEvent transaction,
            Instant evaluatedAt
    ) {
        this.transactionId = transactionId;
        this.blacklistHit = blacklistHit;
        this.reason = reason;
        this.decisionHint = decisionHint;
        this.transaction = transaction;
        this.evaluatedAt = evaluatedAt;
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
