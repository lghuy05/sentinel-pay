package com.example.rule_engine.service;

import java.util.List;

import com.example.rule_engine.service.RiskScoringService.RiskLevel;

public class RiskScoreResult {

    private final int riskScore;
    private final RiskLevel riskLevel;
    private final List<String> triggeredRules;

    public RiskScoreResult(int riskScore, RiskLevel riskLevel, List<String> triggeredRules) {
        this.riskScore = riskScore;
        this.riskLevel = riskLevel;
        this.triggeredRules = triggeredRules;
    }

    public int getRiskScore() {
        return riskScore;
    }

    public RiskLevel getRiskLevel() {
        return riskLevel;
    }

    public List<String> getTriggeredRules() {
        return triggeredRules;
    }
}
