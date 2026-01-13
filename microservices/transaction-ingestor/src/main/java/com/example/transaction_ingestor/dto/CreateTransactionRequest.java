package com.example.transaction_ingestor.dto;

import java.math.BigDecimal;
import java.time.Instant;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;
import jakarta.validation.constraints.Size;

public class CreateTransactionRequest {

    @NotBlank
    private String transactionId;
    
    @NotNull
    private TransactionType type;

    @NotNull
    private Long senderUserId;

    @NotNull
    @Positive
    private BigDecimal amount;

    @NotBlank
    @Size(min = 3, max = 3)
    private String currency; //VND, USD

    @NotBlank
    private String ip;

    @NotNull
    private Instant timestamp;

    private Long merchantId;

    private Long receiverUserId;

    // @NotBlank
    private String deviceId;

}
