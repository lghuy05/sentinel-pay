package com.example.rule_engine.service;

import java.util.ArrayList;
import java.util.List;

import com.example.rule_engine.entity.FraudRule;
import com.example.rule_engine.entity.RuleType;

final class DefaultRuleSet {

    private DefaultRuleSet() {
    }

    static List<FraudRule> buildDefaults() {
        List<FraudRule> defaults = new ArrayList<>();

        defaults.add(buildRule("velocity-1m", RuleType.VELOCITY_1M, 5, 0.4));
        defaults.add(buildRule("amount-1h", RuleType.AMOUNT_1H, 5_000_000, 0.3));
        defaults.add(buildRule("micro-tx-burst-1m", RuleType.MICRO_TX_BURST_1M, 6, 0.6));
        defaults.add(buildRule("small-amount-spread-24h", RuleType.SMALL_VALUE_SPREAD_24H, 30, 0.5));
        defaults.add(buildRule("receiver-inbound-spike-24h", RuleType.RECEIVER_INBOUND_SPIKE_24H, 25, 0.4));
        defaults.add(buildRule("sender-receiver-repeat-24h", RuleType.SENDER_RECEIVER_REPEAT_24H, 3, 0.45));
        defaults.add(buildRule("overseas-high-amount", RuleType.OVERSEAS_HIGH_AMOUNT, 1000, 0.6));
        defaults.add(buildRule("night-new-device", RuleType.NIGHT_NEW_DEVICE, 500_000, 0.5));
        defaults.add(buildRule("absolute-high-amount", RuleType.ABSOLUTE_HIGH_AMOUNT, 200_000_000, 1.2));
        defaults.add(buildRule("cross-border-high-tier", RuleType.CROSS_BORDER_AMOUNT_TIER, 3, 0.7));
        defaults.add(buildRule("cross-border-high-amount", RuleType.CROSS_BORDER_HIGH_AMOUNT, 3000, 0.6));
        defaults.add(buildRule("new-receiver-high-amount", RuleType.NEW_RECEIVER_HIGH_AMOUNT, 1000, 0.5));
        defaults.add(buildRule("first-time-receiver-high-amount", RuleType.FIRST_TIME_RECEIVER_HIGH_AMOUNT, 1000, 0.5));

        return defaults;
    }

    private static FraudRule buildRule(String name, RuleType type, double threshold, double weight) {
        FraudRule rule = new FraudRule();
        rule.setName(name);
        rule.setType(type);
        rule.setThreshold(threshold);
        rule.setWeight(weight);
        rule.setEnabled(true);
        return rule;
    }
}
