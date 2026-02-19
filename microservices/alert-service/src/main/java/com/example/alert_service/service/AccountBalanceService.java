package com.example.alert_service.service;

import java.math.BigDecimal;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestClient;

import com.example.alert_service.event.FraudDecision;
import com.example.alert_service.event.FraudFinalDecisionEvent;

@Service
public class AccountBalanceService {

  private static final Logger log = LoggerFactory.getLogger(AccountBalanceService.class);

  private final RestClient restClient;

  public AccountBalanceService(
      RestClient.Builder builder,
      @Value("${account.service.url:http://localhost:8087}") String baseUrl) {
    this.restClient = builder.baseUrl(baseUrl).build();
  }

  public void applyIfAllowed(FraudFinalDecisionEvent event) {
    if (event == null || event.getFinalDecision() != FraudDecision.ALLOW) {
      return;
    }
    Long senderUserId = event.getSenderUserId();
    Long receiverUserId = event.getReceiverUserId();
    BigDecimal amount = event.getAmount();
    String currency = event.getCurrency();

    if (senderUserId == null || amount == null || currency == null) {
      return;
    }

    long minorAmount = amount.longValue();
    if (minorAmount <= 0) {
      return;
    }

    postBalance(senderUserId, "debit", minorAmount, currency);
    if (receiverUserId != null) {
      postBalance(receiverUserId, "topup", minorAmount, currency);
    }
  }

  private void postBalance(Long userId, String action, long amount, String currency) {
    try {
      restClient.post()
          .uri("/api/v1/accounts/{userId}/" + action, userId)
          .body(Map.of("amount", amount, "currency", currency))
          .retrieve()
          .toBodilessEntity();
    } catch (Exception ex) {
      log.warn("Account balance update failed for {} {}: {}", userId, action, ex.getMessage());
      throw new IllegalStateException("Balance update failed for " + userId + " " + action, ex);
    }
  }
}
