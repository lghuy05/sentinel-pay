package com.example.alert_service.service;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import com.example.alert_service.event.FraudFinalDecisionEvent;

@Service
public class NotificationService {

    private static final Logger log = LoggerFactory.getLogger(NotificationService.class);

    public void sendAlert(FraudFinalDecisionEvent event) {
        if (event.getDecision() == null) {
            return;
        }
        switch (event.getDecision()) {
            case BLOCK:
            case HOLD:
                log.warn("ALERT decision={} txId={} score={}",
                        event.getDecision(),
                        event.getTransactionId(),
                        event.getFinalScore());
                break;
            default:
                log.info("Audit decision={} txId={}", event.getDecision(), event.getTransactionId());
                break;
        }
    }
}
