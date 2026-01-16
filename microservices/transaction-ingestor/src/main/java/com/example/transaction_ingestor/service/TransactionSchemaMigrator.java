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
    }
}
