package com.example.alert_service.service;

import java.time.Instant;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.alert_service.entity.FraudDecisionRecord;
import com.example.alert_service.event.FraudFinalDecisionEvent;
import com.example.alert_service.repository.FraudDecisionRepository;
import com.fasterxml.jackson.databind.ObjectMapper;

@Service
public class FraudDecisionService {

    private final FraudDecisionRepository repository;
    private final NotificationService notificationService;
    private final AccountBalanceService accountBalanceService;
    private final ObjectMapper objectMapper;

    public FraudDecisionService(
            FraudDecisionRepository repository,
            NotificationService notificationService,
            AccountBalanceService accountBalanceService,
            ObjectMapper objectMapper
    ) {
        this.repository = repository;
        this.notificationService = notificationService;
        this.accountBalanceService = accountBalanceService;
        this.objectMapper = objectMapper;
    }

    @Transactional
    public void handleDecision(FraudFinalDecisionEvent event) {
        FraudDecisionRecord record = new FraudDecisionRecord();
        record.setTransactionId(event.getTransactionId());
        record.setDecision(event.getDecision());
        record.setRuleScore(event.getRuleScore());
        record.setBlacklistScore(event.getBlacklistScore());
        record.setMlScore(event.getMlScore());
        record.setFinalScore(event.getFinalScore());
        record.setRuleMatches(serialize(event.getRuleMatches()));
        record.setBlacklistMatches(serialize(event.getBlacklistMatches()));
        record.setRiskScore(event.getRiskScore());
        record.setRiskLevel(event.getRiskLevel());
        record.setTriggeredRules(serialize(event.getTriggeredRules()));
        record.setHardStopMatches(serialize(event.getHardStopMatches()));
        record.setHardStopDecision(event.getHardStopDecision());
        record.setDecidedAt(event.getDecidedAt() != null ? event.getDecidedAt() : Instant.now());

        repository.save(record);
        notificationService.sendAlert(event);
        accountBalanceService.applyIfAllowed(event);
    }

    private String serialize(Object value) {
        try {
            return objectMapper.writeValueAsString(value);
        } catch (Exception e) {
            return "[]";
        }
    }
}
