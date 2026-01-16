package com.example.transaction_ingestor.event;

import java.math.BigDecimal;
import java.time.Instant;

import com.example.transaction_ingestor.dto.TransactionType;

public class TransactionReceivedEvent {

    private final String transactionId;
    private final TransactionType type;

    private final Long senderUserId;
    private final Long receiverUserId;
    private final Long merchantId;

    private final BigDecimal amount;
    private final String currency;
    private final String deviceId;

    private final Instant eventTime;
    private final Instant receivedAt;

    public TransactionReceivedEvent(
            String transactionId,
            TransactionType type,
            Long senderUserId,
            Long receiverUserId,
            Long merchantId,
            BigDecimal amount,
            String currency,
            String deviceId,
            Instant eventTime,
            Instant receivedAt
    ) {
        this.transactionId = transactionId;
        this.type = type;
        this.senderUserId = senderUserId;
        this.receiverUserId = receiverUserId;
        this.merchantId = merchantId;
        this.amount = amount;
        this.currency = currency;
        this.deviceId = deviceId;
        this.eventTime = eventTime;
        this.receivedAt = receivedAt;
    }

    public String getTransactionId() {
        return transactionId;
    }

    public TransactionType getType() {
        return type;
    }

    public Long getSenderUserId() {
        return senderUserId;
    }

    public Long getReceiverUserId() {
        return receiverUserId;
    }

    public Long getMerchantId() {
        return merchantId;
    }

    public BigDecimal getAmount() {
        return amount;
    }

    public String getCurrency() {
        return currency;
    }

    public String getDeviceId() {
        return deviceId;
    }

    public Instant getEventTime() {
        return eventTime;
    }

    public Instant getReceivedAt() {
        return receivedAt;
    }
}
