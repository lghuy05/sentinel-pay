package com.example.rule_engine.service;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import org.springframework.stereotype.Service;

import com.example.rule_engine.event.RuleEvaluationEvent;
import com.example.rule_engine.event.TransactionEnrichedEvent;
import com.fasterxml.jackson.databind.ObjectMapper;

@Service
public class RuleEngineService {

    private final ObjectMapper objectMapper;

    public RuleEngineService(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    public RuleEvaluationEvent evaluate(TransactionEnrichedEvent event) {
        List<String> matched = new ArrayList<>();
        double totalScore = 0.0;
        double amountVnd = toVnd(event.getAmount(), event.getCurrency());

        // Hard rules
        if (event.getTxCountLast1Min() >= 7) {
            matched.add("hard_burst_attack");
            return buildResult(event, 1.0, "RISK", "BLOCK", matched);
        }

        if (isImpossibleGeography(event)) {
            matched.add("hard_impossible_geography");
            return buildResult(event, 1.0, "RISK", "BLOCK", matched);
        }

        // Group 1 — Amount rules
        if (amountVnd > 10_000_000) {
            totalScore += 0.20;
            matched.add("amount_gt_10m");
        }
        if (amountVnd > 50_000_000) {
            totalScore += 0.40;
            matched.add("amount_gt_50m");
        }
        if (amountVnd > 200_000_000) {
            totalScore += 0.50;
            matched.add("amount_gt_200m");
        }
        if (amountVnd > 1_000_000_000) {
            totalScore += 0.70;
            matched.add("amount_gt_1b");
        }
        if (event.getSenderAccountAgeDays() < 7 && amountVnd > 5_000_000) {
            totalScore += 0.35;
            matched.add("new_account_high_amount");
        }

        // Group 2 — Velocity rules
        if (event.getTxCountLast1Min() >= 3) {
            totalScore += 0.25;
            matched.add("velocity_1m_3");
        }
        if (event.getTxCountLast1Min() >= 5) {
            totalScore += 0.45;
            matched.add("velocity_1m_5");
        }
        if (event.getTxCountLast1Hour() >= 10) {
            totalScore += 0.20;
            matched.add("velocity_1h_10");
        }
        if (event.getTxCountLast1Hour() >= 20) {
            totalScore += 0.40;
            matched.add("velocity_1h_20");
        }
        if (event.getAmountLast1Hour() > 20_000_000) {
            totalScore += 0.25;
            matched.add("amount_1h_gt_20m");
        }
        if (event.getAmountLast1Hour() > 50_000_000) {
            totalScore += 0.45;
            matched.add("amount_1h_gt_50m");
        }

        // Group 3 — Geo / overseas rules
        boolean isOverseas = isOverseas(event);
        if (isOverseas && amountVnd > 2_000_000) {
            totalScore += 0.30;
            matched.add("overseas_amount_gt_2m");
        }
        if (isOverseas && amountVnd > 200_000_000) {
            totalScore += 0.60;
            matched.add("overseas_amount_gt_200m");
        }
        if (isOverseas && event.getSenderAccountAgeDays() < 30) {
            totalScore += 0.35;
            matched.add("overseas_new_account");
        }
        if (isOverseas && safeLong(event.getReceiverFirstSeenDays()) < 7) {
            totalScore += 0.30;
            matched.add("overseas_new_receiver");
        }

        // Group 4 — Device rules
        if (event.isNewDevice() && amountVnd > 5_000_000) {
            totalScore += 0.25;
            matched.add("new_device_high_amount");
        }
        if (event.isNewDevice() && event.getTxCountLast1Hour() >= 5) {
            totalScore += 0.35;
            matched.add("new_device_velocity_1h");
        }

        // Group 5 — Account age rules
        if (event.getSenderAccountAgeDays() < 3 && amountVnd > 2_000_000) {
            totalScore += 0.40;
            matched.add("very_new_account_amount");
        }
        if (event.getSenderAccountAgeDays() < 7 && event.getTxCountLast1Hour() >= 5) {
            totalScore += 0.35;
            matched.add("new_account_velocity_1h");
        }

        // Group 6 — Behavior deviation rules
        double avgAmount7d = safeDouble(event.getAvgAmount7d());
        double avgTxPerDay = safeDouble(event.getAvgTxPerDay());
        if (avgAmount7d > 0 && amountVnd > 5 * avgAmount7d) {
            totalScore += 0.30;
            matched.add("amount_gt_5x_avg7d");
        }
        if (avgAmount7d > 0 && amountVnd > 10 * avgAmount7d) {
            totalScore += 0.45;
            matched.add("amount_gt_10x_avg7d");
        }
        if (avgTxPerDay > 0 && event.getTxCountLast1Hour() > 3 * avgTxPerDay) {
            totalScore += 0.30;
            matched.add("velocity_gt_3x_avg");
        }

        // Group 7 — P2P rules
        if (event.getType() != null && event.getType().name().equalsIgnoreCase("P2P_TRANSFER")) {
            if (amountVnd > 10_000_000) {
                totalScore += 0.25;
                matched.add("p2p_high_amount");
            }
            if (isOverseas) {
                totalScore += 0.35;
                matched.add("p2p_overseas");
            }
        }

        double score = Math.min(totalScore, 1.0);
        String band = score < 0.40 ? "SAFE" : (score <= 0.85 ? "GRAY" : "RISK");
        String decisionHint = "SAFE".equals(band) ? "ALLOW" : ("RISK".equals(band) ? "BLOCK" : "GRAY");

        return buildResult(event, score, band, decisionHint, matched);
    }

    private RuleEvaluationEvent buildResult(
            TransactionEnrichedEvent event,
            double score,
            String band,
            String decisionHint,
            List<String> matched
    ) {
        return new RuleEvaluationEvent(
                event.getTransactionId(),
                event.getSenderUserId(),
                event.getReceiverUserId(),
                event.getMerchantId(),
                event.getAmount(),
                event.getCurrency(),
                score,
                band,
                matched,
                decisionHint,
                null,
                buildFeatureMap(event),
                Instant.now()
        );
    }

    private double toVnd(BigDecimal amount, String currency) {
        if (amount == null || currency == null) {
            return 0.0;
        }
        double fx = switch (currency.toUpperCase()) {
            case "USD" -> 25000.0;
            case "EUR" -> 27000.0;
            case "SGD" -> 18000.0;
            case "VND" -> 1.0;
            default -> 1.0;
        };
        return amount.doubleValue() * fx;
    }

    private boolean isOverseas(TransactionEnrichedEvent event) {
        String sender = event.getSenderCountry();
        String receiver = event.getReceiverCountry();
        if (sender == null || receiver == null) {
            return false;
        }
        return !sender.equalsIgnoreCase(receiver);
    }

    private boolean isImpossibleGeography(TransactionEnrichedEvent event) {
        String sender = event.getSenderCountry();
        String last = event.getLastCountry();
        Double distance = event.getGeoDistanceKm();
        long seconds = safeLong(event.getTimeSinceLastTxSeconds());
        if (sender == null || last == null || distance == null) {
            return false;
        }
        return !sender.equalsIgnoreCase(last) && seconds < 600 && distance > 2000.0;
    }

    private double safeDouble(Double value) {
        return value == null ? 0.0 : value;
    }

    private long safeLong(Long value) {
        return value == null ? 0L : value;
    }

    private Map<String, Object> buildFeatureMap(TransactionEnrichedEvent event) {
        return objectMapper.convertValue(event, Map.class);
    }
}
