package com.example.rule_engine.service;

import java.time.Instant;
import java.time.ZoneOffset;
import java.util.ArrayList;
import java.util.List;

import jakarta.annotation.PostConstruct;

import org.springframework.stereotype.Service;
import org.springframework.context.annotation.DependsOn;
import org.springframework.transaction.annotation.Transactional;

import com.example.rule_engine.dto.AmountRiskTier;
import com.example.rule_engine.entity.FraudRule;
import com.example.rule_engine.entity.RuleType;
import com.example.rule_engine.event.RuleEvaluationEvent;
import com.example.rule_engine.event.TransactionEnrichedEvent;
import com.example.rule_engine.repository.FraudRuleRepository;

@Service
@DependsOn("ruleSchemaMigrator")
public class RuleEngineService {

    private static final double MICRO_AMOUNT_USD = 15.0;
    private static final double SMALL_AMOUNT_AVG_USD_24H = 20.0;

    private final FraudRuleRepository ruleRepository;
    private final RiskScoringService riskScoringService;
    private final List<FraudRule> cachedRules = new ArrayList<>();

    public RuleEngineService(
            FraudRuleRepository ruleRepository,
            RiskScoringService riskScoringService
    ) {
        this.ruleRepository = ruleRepository;
        this.riskScoringService = riskScoringService;
    }

    @PostConstruct
    @Transactional
    public void loadRules() {
        seedDefaults();
        cachedRules.clear();
        cachedRules.addAll(ruleRepository.findAll());
    }

    public RuleEvaluationEvent evaluate(TransactionEnrichedEvent event) {
        List<String> matched = new ArrayList<>();
        double score = 0.0;

        for (FraudRule rule : cachedRules) {
            if (!rule.isEnabled()) {
                continue;
            }
            if (matches(rule, event)) {
                matched.add(rule.getName());
                score += rule.getWeight();
            }
        }

        HardStopResult hardStop = evaluateHardStops(event);
        RiskScoreResult riskScore = riskScoringService.score(event);

        return new RuleEvaluationEvent(
                event.getTransactionId(),
                event.getSenderUserId(),
                event.getReceiverUserId(),
                event.getMerchantId(),
                event.getAmount(),
                event.getCurrency(),
                score,
                matched,
                riskScore.getRiskScore(),
                riskScore.getRiskLevel().name(),
                riskScore.getTriggeredRules(),
                hardStop.matches(),
                hardStop.decision(),
                Instant.now()
        );
    }

    private boolean matches(FraudRule rule, TransactionEnrichedEvent event) {
        switch (rule.getType()) {
            case VELOCITY_1M:
                return event.getTxCountLast1Min() >= rule.getThreshold();
            case AMOUNT_1H:
                return event.getTxAmountLast1Hour() >= rule.getThreshold();
            case MICRO_TX_BURST_1M:
                return event.getTxCountLast1Min() >= rule.getThreshold()
                        && event.getAmountUsdEquivalent() <= MICRO_AMOUNT_USD;
            case SMALL_VALUE_SPREAD_24H:
                return event.getSenderTxCount24h() >= rule.getThreshold()
                        && averageAmountUsd24h(event) > 0
                        && averageAmountUsd24h(event) <= SMALL_AMOUNT_AVG_USD_24H;
            case RECEIVER_INBOUND_SPIKE_24H:
                return event.getReceiverInboundCount24h() >= rule.getThreshold();
            case SENDER_RECEIVER_REPEAT_24H:
                return event.getSenderReceiverTxCount24h() >= rule.getThreshold();
            case OVERSEAS_HIGH_AMOUNT:
                return event.isOverseas()
                        && event.getAmountUsdEquivalent() >= rule.getThreshold();
            case NIGHT_NEW_DEVICE:
                int hour = event.getEventTime()
                        .atZone(ZoneOffset.UTC)
                        .getHour();
                return event.isNewDevice()
                        && (hour >= 0 && hour <= 5)
                        && event.getAmount().doubleValue() >= rule.getThreshold();
            case ABSOLUTE_HIGH_AMOUNT:
                return event.getAmount().doubleValue() >= rule.getThreshold();
            case CROSS_BORDER_AMOUNT_TIER:
                return event.isCrossBorder()
                        && tierRank(event.getAmountRiskTier()) >= (int) rule.getThreshold();
            case CROSS_BORDER_HIGH_AMOUNT:
                return event.isCrossBorder()
                        && event.getAmountUsdEquivalent() >= rule.getThreshold();
            case NEW_RECEIVER_HIGH_AMOUNT:
                return event.getReceiverAccountAgeDays() < 7
                        && event.getAmountUsdEquivalent() >= rule.getThreshold();
            case FIRST_TIME_RECEIVER_HIGH_AMOUNT:
                return event.isFirstTimeContact()
                        && event.getAmountUsdEquivalent() >= rule.getThreshold();
            default:
                return false;
        }
    }

    private void seedDefaults() {
        List<FraudRule> defaults = new ArrayList<>();

        defaults.add(buildRule(
                "velocity-1m",
                RuleType.VELOCITY_1M,
                5,
                0.4
        ));
        defaults.add(buildRule(
                "amount-1h",
                RuleType.AMOUNT_1H,
                5_000_000,
                0.3
        ));
        defaults.add(buildRule(
                "micro-tx-burst-1m",
                RuleType.MICRO_TX_BURST_1M,
                6,
                0.6
        ));
        defaults.add(buildRule(
                "small-amount-spread-24h",
                RuleType.SMALL_VALUE_SPREAD_24H,
                30,
                0.5
        ));
        defaults.add(buildRule(
                "receiver-inbound-spike-24h",
                RuleType.RECEIVER_INBOUND_SPIKE_24H,
                25,
                0.4
        ));
        defaults.add(buildRule(
                "sender-receiver-repeat-24h",
                RuleType.SENDER_RECEIVER_REPEAT_24H,
                3,
                0.45
        ));
        defaults.add(buildRule(
                "overseas-high-amount",
                RuleType.OVERSEAS_HIGH_AMOUNT,
                1000,
                0.6
        ));
        defaults.add(buildRule(
                "night-new-device",
                RuleType.NIGHT_NEW_DEVICE,
                500_000,
                0.5
        ));
        defaults.add(buildRule(
                "absolute-high-amount",
                RuleType.ABSOLUTE_HIGH_AMOUNT,
                200_000_000,
                1.2
        ));
        defaults.add(buildRule(
                "cross-border-high-tier",
                RuleType.CROSS_BORDER_AMOUNT_TIER,
                3,
                0.7
        ));
        defaults.add(buildRule(
                "cross-border-high-amount",
                RuleType.CROSS_BORDER_HIGH_AMOUNT,
                3000,
                0.6
        ));
        defaults.add(buildRule(
                "new-receiver-high-amount",
                RuleType.NEW_RECEIVER_HIGH_AMOUNT,
                1000,
                0.5
        ));
        defaults.add(buildRule(
                "first-time-receiver-high-amount",
                RuleType.FIRST_TIME_RECEIVER_HIGH_AMOUNT,
                1000,
                0.5
        ));

        for (FraudRule rule : defaults) {
            ruleRepository.findByName(rule.getName())
                    .ifPresentOrElse(existing -> {
                    }, () -> ruleRepository.save(rule));
        }
    }

    private FraudRule buildRule(
            String name,
            RuleType type,
            double threshold,
            double weight
    ) {
        FraudRule rule = new FraudRule();
        rule.setName(name);
        rule.setType(type);
        rule.setThreshold(threshold);
        rule.setWeight(weight);
        rule.setEnabled(true);
        return rule;
    }

    private int tierRank(AmountRiskTier tier) {
        if (tier == null) {
            return 0;
        }
        switch (tier) {
            case LOW:
                return 1;
            case MEDIUM:
                return 2;
            case HIGH:
                return 3;
            case CRITICAL:
                return 4;
            default:
                return 0;
        }
    }

    private HardStopResult evaluateHardStops(TransactionEnrichedEvent event) {
        List<String> matches = new ArrayList<>();
        String decision = null;

        if (event.getAmountUsdEquivalent() >= 50_000) {
            matches.add("HARD_MAX_AMOUNT_BLOCK");
            decision = "BLOCK";
        } else if (event.getAmountUsdEquivalent() >= 10_000) {
            matches.add("HARD_LARGE_AMOUNT_REVIEW");
            decision = "REVIEW";
        }

        if (event.isCrossBorder() && event.getAmountUsdEquivalent() >= 3_000) {
            matches.add("HARD_CROSS_BORDER_LARGE_REVIEW");
            decision = pickHigher(decision, "REVIEW");
        }

        if (event.getReceiverAccountAgeDays() < 7 && event.getAmountUsdEquivalent() >= 1_000) {
            matches.add("HARD_NEW_RECEIVER_LARGE_REVIEW");
            decision = pickHigher(decision, "REVIEW");
        }

        if (event.isFirstTimeContact() && event.getAmountUsdEquivalent() >= 1_000) {
            matches.add("HARD_FIRST_TIME_RECEIVER_LARGE_REVIEW");
            decision = pickHigher(decision, "REVIEW");
        }

        if (event.getSenderTxCount24h() >= 50) {
            matches.add("HARD_RATE_LIMIT_TX_COUNT");
            decision = pickHigher(decision, "REVIEW");
        }

        if (event.getSenderTotalAmountUsd24h() >= 20_000) {
            matches.add("HARD_RATE_LIMIT_AMOUNT");
            decision = pickHigher(decision, "REVIEW");
        }

        if (event.getSenderTxCount24h() >= 60 && averageAmountUsd24h(event) <= 10.0) {
            matches.add("HARD_MICRO_BURST_REVIEW");
            decision = pickHigher(decision, "REVIEW");
        }

        if (event.getReceiverInboundCount24h() >= 40) {
            matches.add("HARD_RECEIVER_INBOUND_SPIKE");
            decision = pickHigher(decision, "REVIEW");
        }

        if (event.getSenderReceiverTxCount24h() >= 5 && averageAmountUsd24h(event) <= 20.0) {
            matches.add("HARD_REPEAT_RECEIVER_SMALL_VALUE");
            decision = pickHigher(decision, "REVIEW");
        }

        return new HardStopResult(matches, decision);
    }

    private double averageAmountUsd24h(TransactionEnrichedEvent event) {
        if (event.getSenderTxCount24h() <= 0) {
            return 0.0;
        }
        return event.getSenderTotalAmountUsd24h() / event.getSenderTxCount24h();
    }

    private String pickHigher(String current, String candidate) {
        if (current == null) {
            return candidate;
        }
        if ("BLOCK".equalsIgnoreCase(current)) {
            return current;
        }
        if ("BLOCK".equalsIgnoreCase(candidate)) {
            return "BLOCK";
        }
        return "REVIEW";
    }

    private record HardStopResult(List<String> matches, String decision) {}
}
