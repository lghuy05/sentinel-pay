package com.example.alert_service.service;

import java.time.Instant;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.alert_service.entity.FraudDecisionRecord;
import com.example.alert_service.event.FraudDecision;
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
        if (event == null || event.getFinalDecision() == null) {
            return;
        }
        if (event.getFinalDecision() == FraudDecision.ALLOW) {
            accountBalanceService.applyIfAllowed(event);
            return;
        }

        FraudDecisionRecord record = new FraudDecisionRecord();
        record.setTransactionId(event.getTransactionId());
        record.setDecision(event.getFinalDecision());
        record.setDecisionReason(event.getDecisionReason());
        record.setPayloadJson(serialize(event));
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
