package com.example.account_service.entity;

import java.time.Instant;

import com.example.account_service.dto.AccountStatus;
import com.example.account_service.dto.KycLevel;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.Id;
import jakarta.persistence.Table;

import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
@Entity
@Table(name = "accounts")
public class Account {

    // TESTING ONLY: This service is designed for demo and QA flows, not production.
    @Id
    private Long userId;

    @Column(nullable = false, length = 2)
    private String accountCountry;

    @Column(nullable = false, length = 3)
    private String homeCurrency;

    @Column(nullable = false)
    private Instant createdAt = Instant.now();

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 10)
    private KycLevel kycLevel = KycLevel.BASIC;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 10)
    private AccountStatus status = AccountStatus.ACTIVE;

    @Column(nullable = false)
    private long balanceMinor;
}
