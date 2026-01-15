package com.example.rule_engine.event;

import java.time.Instant;
import java.util.List;

public class RuleEvaluationEvent {

    private String transactionId;
    private double ruleScore;
    private List<String> matchedRules;
    private Instant evaluatedAt;

    public RuleEvaluationEvent() {
    }

    public RuleEvaluationEvent(
            String transactionId,
            double ruleScore,
            List<String> matchedRules,
            Instant evaluatedAt
    ) {
        this.transactionId = transactionId;
        this.ruleScore = ruleScore;
        this.matchedRules = matchedRules;
        this.evaluatedAt = evaluatedAt;
    }

    public String getTransactionId() {
        return transactionId;
    }

    public void setTransactionId(String transactionId) {
        this.transactionId = transactionId;
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

    public Instant getEvaluatedAt() {
        return evaluatedAt;
    }

    public void setEvaluatedAt(Instant evaluatedAt) {
        this.evaluatedAt = evaluatedAt;
    }
}
