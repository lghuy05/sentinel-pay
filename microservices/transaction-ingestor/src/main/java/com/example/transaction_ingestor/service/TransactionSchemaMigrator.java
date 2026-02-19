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

        try {
            jdbcTemplate.execute("ALTER TABLE outbox_event ALTER COLUMN payload TYPE TEXT");
            jdbcTemplate.execute("ALTER TABLE outbox_event ALTER COLUMN last_error TYPE TEXT");
            log.info("Ensured outbox_event payload and last_error are TEXT");
        } catch (Exception e) {
            log.warn("Failed to update outbox_event column types", e);
        }
    }
}
