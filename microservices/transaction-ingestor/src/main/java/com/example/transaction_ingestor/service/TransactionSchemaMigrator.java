package com.example.transaction_ingestor.service;

import jakarta.annotation.PostConstruct;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Component;

@Component
public class TransactionSchemaMigrator {

    private static final Logger log = LoggerFactory.getLogger(TransactionSchemaMigrator.class);

    private final JdbcTemplate jdbcTemplate;

    public TransactionSchemaMigrator(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @PostConstruct
    public void dropLegacyIpColumn() {
        try {
            jdbcTemplate.execute("ALTER TABLE transaction_records DROP COLUMN IF EXISTS ip");
            log.info("Ensured legacy ip column is removed from transaction_records");
        } catch (Exception e) {
            log.warn("Failed to drop legacy ip column", e);
        }

        try {
            jdbcTemplate.execute(
                    "CREATE INDEX IF NOT EXISTS idx_transaction_records_received_at " +
                    "ON transaction_records (received_at DESC)"
            );
            jdbcTemplate.execute(
                    "CREATE INDEX IF NOT EXISTS idx_transaction_records_sender_time_amount_currency " +
                    "ON transaction_records (sender_user_id, event_time, amount, currency)"
            );
            log.info("Ensured performance indexes exist on transaction_records");
        } catch (Exception e) {
            log.warn("Failed to create performance indexes", e);
        }
    }
}
