package com.example.alert_service.service;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import com.example.alert_service.event.FraudFinalDecisionEvent;

@Service
public class NotificationService {

    private static final Logger log = LoggerFactory.getLogger(NotificationService.class);

    public void sendAlert(FraudFinalDecisionEvent event) {
        if (event.getFinalDecision() == null) {
            return;
        }
        switch (event.getFinalDecision()) {
            case BLOCK:
            case HOLD:
                log.warn("ALERT decision={} txId={} reason={}",
                        event.getFinalDecision(),
                        event.getTransactionId(),
                        event.getDecisionReason());
                break;
            default:
                log.info("Audit decision={} txId={}", event.getFinalDecision(), event.getTransactionId());
                break;
        }
    }
}
