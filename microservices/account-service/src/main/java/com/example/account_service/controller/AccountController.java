package com.example.account_service.controller;

import java.util.List;

import jakarta.validation.Valid;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import com.example.account_service.dto.AccountResponse;
import com.example.account_service.dto.BalanceRequest;
import com.example.account_service.dto.CreateAccountRequest;
import com.example.account_service.dto.UpdateAccountRequest;
import com.example.account_service.service.AccountService;

@RestController
@RequestMapping("/api/v1/accounts")
public class AccountController {

    private final AccountService accountService;

    public AccountController(AccountService accountService) {
        this.accountService = accountService;
    }

    @PostMapping
    public ResponseEntity<AccountResponse> create(@Valid @RequestBody CreateAccountRequest request) {
        return ResponseEntity.ok(accountService.create(request));
    }

    @GetMapping("/{userId}")
    public ResponseEntity<AccountResponse> get(@PathVariable Long userId) {
        return ResponseEntity.ok(accountService.get(userId));
    }

    @GetMapping
    public ResponseEntity<List<AccountResponse>> list(
            @RequestParam(name = "limit", defaultValue = "50") int limit,
            @RequestParam(name = "offset", defaultValue = "0") int offset
    ) {
        return ResponseEntity.ok(accountService.list(limit, offset));
    }

    @PatchMapping("/{userId}")
    public ResponseEntity<AccountResponse> update(
            @PathVariable Long userId,
            @RequestBody UpdateAccountRequest request
    ) {
        return ResponseEntity.ok(accountService.update(userId, request));
    }

    @PostMapping("/{userId}/topup")
    public ResponseEntity<AccountResponse> topup(
            @PathVariable Long userId,
            @Valid @RequestBody BalanceRequest request
    ) {
        return ResponseEntity.ok(accountService.topup(userId, request));
    }

    @PostMapping("/{userId}/debit")
    public ResponseEntity<AccountResponse> debit(
            @PathVariable Long userId,
            @Valid @RequestBody BalanceRequest request
    ) {
        return ResponseEntity.ok(accountService.debit(userId, request));
    }

    @DeleteMapping("/{userId}")
    public ResponseEntity<Void> delete(@PathVariable Long userId) {
        accountService.delete(userId);
        return ResponseEntity.noContent().build();
    }
}
