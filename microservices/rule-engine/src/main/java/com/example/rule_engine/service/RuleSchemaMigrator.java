package com.example.rule_engine.service;

import java.util.Arrays;
import java.util.stream.Collectors;

import jakarta.annotation.PostConstruct;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Component;

import com.example.rule_engine.entity.RuleType;

@Component
public class RuleSchemaMigrator {

    private static final Logger log = LoggerFactory.getLogger(RuleSchemaMigrator.class);

    private final JdbcTemplate jdbcTemplate;

    public RuleSchemaMigrator(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @PostConstruct
    public void updateRuleTypeConstraint() {
        String constraintName = "fraud_rules_type_check";
        String allowed = Arrays.stream(RuleType.values())
                .map(RuleType::name)
                .map(value -> "'" + value + "'")
                .collect(Collectors.joining(", "));
        String constraintSql = "CHECK (type in (" + allowed + "))";

        try {
            jdbcTemplate.execute("ALTER TABLE fraud_rules DROP CONSTRAINT IF EXISTS " + constraintName);
            jdbcTemplate.execute("ALTER TABLE fraud_rules ADD CONSTRAINT " + constraintName + " " + constraintSql);
            log.info("Updated fraud_rules type constraint to include: {}", allowed);
        } catch (Exception e) {
            log.warn("Failed to update fraud_rules type constraint", e);
        }
    }
}
