package com.example.fraud_orchestrator.controller;

import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;
import org.springframework.http.HttpStatus;

import com.example.fraud_orchestrator.entity.FraudDecisionRecord;
import com.example.fraud_orchestrator.repository.FraudDecisionRepository;

@RestController
public class FeedbackController {

    private final FraudDecisionRepository repository;

    public FeedbackController(FraudDecisionRepository repository) {
        this.repository = repository;
    }

    @PostMapping("/feedback")
    public ResponseEntity<Void> submit(@Valid @RequestBody FeedbackRequest request) {
        FraudDecisionRecord record = repository.findByTransactionId(request.transactionId)
                .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "Decision not found"));
        record.setTrueLabel(request.label == 1);
        record.setReviewed(true);
        repository.save(record);
        return ResponseEntity.accepted().build();
    }

    public static final class FeedbackRequest {
        @NotBlank
        private String transactionId;
        @NotNull
        private Integer label;

        public String getTransactionId() {
            return transactionId;
        }

        public void setTransactionId(String transactionId) {
            this.transactionId = transactionId;
        }

        public Integer getLabel() {
            return label;
        }

        public void setLabel(Integer label) {
            this.label = label;
        }
    }
}
