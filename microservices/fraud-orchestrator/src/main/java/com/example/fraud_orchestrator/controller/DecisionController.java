package com.example.fraud_orchestrator.controller;

import java.util.List;

import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Sort;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;

import com.example.fraud_orchestrator.entity.FraudDecisionRecord;
import com.example.fraud_orchestrator.repository.FraudDecisionRepository;

@RestController
@RequestMapping("/decisions")
public class DecisionController {

    private final FraudDecisionRepository repository;

    public DecisionController(FraudDecisionRepository repository) {
        this.repository = repository;
    }

    @GetMapping
    public ResponseEntity<List<FraudDecisionRecord>> list(
            @RequestParam(name = "limit", defaultValue = "50") int limit,
            @RequestParam(name = "reviewed", required = false) Boolean reviewed
    ) {
        int size = Math.max(1, Math.min(limit, 200));
        if (Boolean.FALSE.equals(reviewed)) {
            List<FraudDecisionRecord> records = repository
                    .findByReviewedFalse(PageRequest.of(0, size, Sort.by("createdAt").descending()))
                    .getContent();
            return ResponseEntity.ok(records);
        }
        List<FraudDecisionRecord> records = repository
                .findAll(PageRequest.of(0, size, Sort.by("createdAt").descending()))
                .getContent();
        return ResponseEntity.ok(records);
    }

    @GetMapping("/{transactionId}")
    public ResponseEntity<FraudDecisionRecord> get(@PathVariable String transactionId) {
        FraudDecisionRecord record = repository.findByTransactionId(transactionId)
                .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "Decision not found"));
        return ResponseEntity.ok(record);
    }
}
