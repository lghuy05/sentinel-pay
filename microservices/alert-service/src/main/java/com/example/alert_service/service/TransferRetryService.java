package com.example.alert_service.service;

import java.util.List;

import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import com.example.alert_service.entity.Transfer;
import com.example.alert_service.entity.TransferStatus;
import com.example.alert_service.repository.TransferRepository;

@Service
public class TransferRetryService {
  private final TransferRepository transferRepository;
  private final FraudDecisionService fraudDecisionService;

  public TransferRetryService(TransferRepository transferRepository,
      FraudDecisionService fraudDecisionService) {
    this.transferRepository = transferRepository;
    this.fraudDecisionService = fraudDecisionService;
  }

  @Scheduled(fixedDelayString = "${transfer.retry.delay-ms:5000}")
  public void retryFailedTransfers() {
    List<Transfer> failed = transferRepository.findTop100ByStatusOrderByUpdatedAtAsc(TransferStatus.FAILED_RETRYABLE);
    for (Transfer t : failed) {
      fraudDecisionService.retryTransfer(t);
    }
  }
}
