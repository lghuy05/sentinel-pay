package com.example.rule_engine.service;

import java.time.Instant;
import java.time.ZoneOffset;
import java.util.ArrayList;
import java.util.List;

import jakarta.annotation.PostConstruct;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.rule_engine.entity.FraudRule;
import com.example.rule_engine.entity.RuleType;
import com.example.rule_engine.event.RuleEvaluationEvent;
import com.example.rule_engine.event.TransactionEnrichedEvent;
import com.example.rule_engine.repository.FraudRuleRepository;

@Service
public class RuleEngineService {

    private final FraudRuleRepository ruleRepository;
    private final List<FraudRule> cachedRules = new ArrayList<>();

    public RuleEngineService(FraudRuleRepository ruleRepository) {
        this.ruleRepository = ruleRepository;
    }

    @PostConstruct
    @Transactional
    public void loadRules() {
        if (ruleRepository.count() == 0) {
            seedDefaults();
        }
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

        return new RuleEvaluationEvent(
                event.getTransactionId(),
                score,
                matched,
                Instant.now()
        );
    }

    private boolean matches(FraudRule rule, TransactionEnrichedEvent event) {
        switch (rule.getType()) {
            case VELOCITY_1M:
                return event.getTxCountLast1Min() >= rule.getThreshold();
            case AMOUNT_1H:
                return event.getTxAmountLast1Hour() >= rule.getThreshold();
            case OVERSEAS_HIGH_AMOUNT:
                return !"VND".equalsIgnoreCase(event.getCurrency())
                        && event.getAmount().doubleValue() >= rule.getThreshold();
            case NIGHT_NEW_DEVICE:
                int hour = event.getEventTime()
                        .atZone(ZoneOffset.UTC)
                        .getHour();
                return event.isNewDevice()
                        && (hour >= 0 && hour <= 5)
                        && event.getAmount().doubleValue() >= rule.getThreshold();
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
                "overseas-high-amount",
                RuleType.OVERSEAS_HIGH_AMOUNT,
                1_000_000,
                0.6
        ));
        defaults.add(buildRule(
                "night-new-device",
                RuleType.NIGHT_NEW_DEVICE,
                500_000,
                0.5
        ));

        ruleRepository.saveAll(defaults);
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
}
