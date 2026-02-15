package com.example.transaction_ingestor.controller;

import jakarta.validation.Valid;

import java.util.List;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import com.example.transaction_ingestor.dto.CreateTransactionRequest;
import com.example.transaction_ingestor.entity.TransactionRecord;
import com.example.transaction_ingestor.service.TransactionIngestService;

@RestController
@RequestMapping("/api/v1/transactions")
public class TransactionController {

  private final TransactionIngestService ingestService;

  public TransactionController(TransactionIngestService ingestService) {
    this.ingestService = ingestService;
  }

  @PostMapping
  public ResponseEntity<TransactionRecord> ingest(
      @Valid @RequestBody CreateTransactionRequest dto) {
    TransactionRecord record = ingestService.ingest(dto);
    return ResponseEntity.accepted().body(record);
  }

  @GetMapping
  public ResponseEntity<List<TransactionRecord>> listRecent(
      @RequestParam(name = "limit", defaultValue = "50") int limit) {
    return ResponseEntity.ok(ingestService.listRecent(limit));
  }
}
