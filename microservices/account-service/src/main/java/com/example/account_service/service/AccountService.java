package com.example.account_service.service;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.ThreadLocalRandom;
import java.math.RoundingMode;

import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.account_service.dto.AccountResponse;
import com.example.account_service.dto.AccountStatus;
import com.example.account_service.dto.BalanceRequest;
import com.example.account_service.dto.CreateAccountRequest;
import com.example.account_service.dto.KycLevel;
import com.example.account_service.dto.UpdateAccountRequest;
import com.example.account_service.entity.Account;
import com.example.account_service.repository.AccountRepository;

@Service
public class AccountService {

    private static final long VND_PER_USD = 25_000L;

    private final AccountRepository repository;

    public AccountService(AccountRepository repository) {
        this.repository = repository;
    }

    @Transactional
    public AccountResponse create(CreateAccountRequest request) {
        Long initialBalance = request.getInitialBalance();
        if (initialBalance != null && initialBalance < 0) {
            throw new IllegalArgumentException("initialBalance must be non-negative");
        }
        Account account = new Account();
        account.setUserId(request.getUserId() != null ? request.getUserId() : generateUserId());
        account.setAccountCountry(request.getAccountCountry());
        account.setHomeCurrency(request.getHomeCurrency());
        account.setCreatedAt(request.getCreatedAt() != null ? request.getCreatedAt() : Instant.now());
        account.setKycLevel(request.getKycLevel() != null ? request.getKycLevel() : KycLevel.BASIC);
        account.setStatus(request.getStatus() != null ? request.getStatus() : AccountStatus.ACTIVE);
        account.setBalanceMinor(Optional.ofNullable(initialBalance).orElse(0L));
        return toResponse(repository.save(account));
    }

    @Transactional(readOnly = true)
    public AccountResponse get(Long userId) {
        Account account = repository.findById(userId)
                .orElseThrow(() -> new IllegalArgumentException("Account not found"));
        return toResponse(account);
    }

    @Transactional(readOnly = true)
    public List<AccountResponse> list(int limit, int offset) {
        int size = Math.max(1, Math.min(limit, 200));
        int page = Math.max(0, offset / size);
        return repository.findAll(PageRequest.of(page, size, Sort.by("userId").ascending()))
                .map(this::toResponse)
                .toList();
    }

    @Transactional
    public AccountResponse update(Long userId, UpdateAccountRequest request) {
        Account account = repository.findById(userId)
                .orElseThrow(() -> new IllegalArgumentException("Account not found"));
        if (request.getAccountCountry() != null) {
            account.setAccountCountry(request.getAccountCountry());
        }
        if (request.getCreatedAt() != null) {
            account.setCreatedAt(request.getCreatedAt());
        }
        if (request.getKycLevel() != null) {
            account.setKycLevel(request.getKycLevel());
        }
        if (request.getStatus() != null) {
            account.setStatus(request.getStatus());
        }
        return toResponse(repository.save(account));
    }

    @Transactional
    public AccountResponse topup(Long userId, BalanceRequest request) {
        if (request.getAmount() <= 0) {
            throw new IllegalArgumentException("amount must be positive");
        }
        Account account = repository.findById(userId)
                .orElseThrow(() -> new IllegalArgumentException("Account not found"));
        long delta = convertToMinor(request.getAmount(), request.getCurrency(), account.getHomeCurrency());
        account.setBalanceMinor(account.getBalanceMinor() + delta);
        return toResponse(repository.save(account));
    }

    @Transactional
    public AccountResponse debit(Long userId, BalanceRequest request) {
        if (request.getAmount() <= 0) {
            throw new IllegalArgumentException("amount must be positive");
        }
        Account account = repository.findById(userId)
                .orElseThrow(() -> new IllegalArgumentException("Account not found"));
        long delta = convertToMinor(request.getAmount(), request.getCurrency(), account.getHomeCurrency());
        if (account.getBalanceMinor() - delta < 0) {
            throw new IllegalArgumentException("Insufficient funds");
        }
        account.setBalanceMinor(account.getBalanceMinor() - delta);
        return toResponse(repository.save(account));
    }

    @Transactional
    public void delete(Long userId) {
        if (!repository.existsById(userId)) {
            throw new IllegalArgumentException("Account not found");
        }
        repository.deleteById(userId);
    }

    private long generateUserId() {
        return ThreadLocalRandom.current().nextLong(100000, 999999);
    }

    private long convertToMinor(Long amount, String currency, String targetCurrency) {
        if (amount == null) {
            return 0L;
        }
        if (currency == null || targetCurrency == null) {
            return amount;
        }
        String from = currency.toUpperCase();
        String to = targetCurrency.toUpperCase();
        if (from.equals(to)) {
            return amount;
        }
        BigDecimal value = BigDecimal.valueOf(amount);
        if ("USD".equals(from) && "VND".equals(to)) {
            return value.multiply(BigDecimal.valueOf(VND_PER_USD)).longValue();
        }
        if ("VND".equals(from) && "USD".equals(to)) {
            return value.divide(BigDecimal.valueOf(VND_PER_USD), 0, RoundingMode.HALF_UP).longValue();
        }
        return value.longValue();
    }

    private AccountResponse toResponse(Account account) {
        AccountResponse response = new AccountResponse();
        response.setUserId(account.getUserId());
        response.setAccountCountry(account.getAccountCountry());
        response.setHomeCurrency(account.getHomeCurrency());
        response.setCreatedAt(account.getCreatedAt());
        response.setKycLevel(account.getKycLevel());
        response.setStatus(account.getStatus());
        response.setBalanceMinor(account.getBalanceMinor());
        return response;
    }
}
