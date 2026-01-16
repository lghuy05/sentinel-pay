package com.example.rule_engine.service;

import java.time.Instant;
import java.time.ZoneOffset;
import java.util.ArrayList;
import java.util.List;

import org.springframework.stereotype.Service;

import com.example.rule_engine.event.TransactionEnrichedEvent;

@Service
public class RiskScoringService {


    public enum RiskLevel {
        LOW,
        MEDIUM,
        HIGH
    }

    public RiskScoreResult score(TransactionEnrichedEvent event) {
        int score = 0;
        List<String> triggered = new ArrayList<>();

        if (event.isCrossBorder()) {
            score += 40;
            triggered.add("cross-border");
        }

        if (event.isCrossBorder() && event.getAmountUsdEquivalent() >= 400) {
            score += 30;
            triggered.add("cross-border-high-amount");
        }

        if (isNightLocal(event.getEventTime(), event.getSenderAccountCountry())) {
            score += 25;
            triggered.add("late-night");
        }

        if (event.getDailyAmountUtilization() >= 0.8) {
            score += 20;
            triggered.add("daily-limit-80pct");
        }

        if (event.isDailyLimitExceeded()) {
            score += 50;
            triggered.add("daily-limit-exceeded");
        }

        if (event.isFirstTimeContact()) {
            score += 30;
            triggered.add("first-time-contact");
        }

        double avgAmountUsd24h = averageAmountUsd24h(event);
        if (event.getTxCountLast1Min() >= 6 && event.getAmountUsdEquivalent() <= 15) {
            score += 35;
            triggered.add("micro-burst-1m");
        }

        if (event.getSenderTxCount24h() >= 30 && avgAmountUsd24h > 0 && avgAmountUsd24h <= 20) {
            score += 25;
            triggered.add("small-amount-spread-24h");
        }

        if (event.getReceiverInboundCount24h() >= 25) {
            score += 20;
            triggered.add("receiver-inbound-spike");
        }

        if (event.getSenderReceiverTxCount24h() >= 3) {
            score += 22;
            triggered.add("repeat-receiver-24h");
        }

        RiskLevel level = score >= 70 ? RiskLevel.HIGH : (score >= 40 ? RiskLevel.MEDIUM : RiskLevel.LOW);
        return new RiskScoreResult(score, level, triggered);
    }

    private boolean isNightLocal(Instant eventTime, String senderCountry) {
        if (eventTime == null) {
            return false;
        }
        ZoneOffset offset = ZoneOffset.UTC;
        if (senderCountry != null) {
            switch (senderCountry.toUpperCase()) {
                case "VN":
                    offset = ZoneOffset.ofHours(7);
                    break;
                case "US":
                    offset = ZoneOffset.ofHours(-5);
                    break;
                case "JP":
                    offset = ZoneOffset.ofHours(9);
                    break;
                case "SG":
                    offset = ZoneOffset.ofHours(8);
                    break;
                case "CA":
                    offset = ZoneOffset.ofHours(-5);
                    break;
                case "MX":
                    offset = ZoneOffset.ofHours(-6);
                    break;
                default:
                    offset = ZoneOffset.UTC;
                    break;
            }
        }
        int hour = eventTime.atOffset(offset).getHour();
        return hour >= 23 || hour <= 4;
    }

    private double averageAmountUsd24h(TransactionEnrichedEvent event) {
        if (event.getSenderTxCount24h() <= 0) {
            return 0.0;
        }
        return event.getSenderTotalAmountUsd24h() / event.getSenderTxCount24h();
    }
}
