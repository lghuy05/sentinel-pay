package com.example.blacklist_service.event;

import java.math.BigDecimal;
import java.time.Instant;

import com.example.blacklist_service.dto.TransactionType;

public class TransactionEnrichedEvent {

    private String transactionId;
    private TransactionType type;
    private Long senderUserId;
    private Long receiverUserId;
    private Long merchantId;
    private BigDecimal amount;
    private String currency;
    private String deviceId;
    private Instant eventTime;
    private Instant receivedAt;
    private long txCountLast1Min;
    private double txAmountLast1Hour;
    private long uniqueMerchantsLast24h;
    private Instant deviceLastSeenAt;
    private boolean newDevice;
    private Instant featureComputedAt;

    public TransactionEnrichedEvent() {
    }

    public String getTransactionId() {
        return transactionId;
    }

    public void setTransactionId(String transactionId) {
        this.transactionId = transactionId;
    }

    public TransactionType getType() {
        return type;
    }

    public void setType(TransactionType type) {
        this.type = type;
    }

    public Long getSenderUserId() {
        return senderUserId;
    }

    public void setSenderUserId(Long senderUserId) {
        this.senderUserId = senderUserId;
    }

    public Long getReceiverUserId() {
        return receiverUserId;
    }

    public void setReceiverUserId(Long receiverUserId) {
        this.receiverUserId = receiverUserId;
    }

    public Long getMerchantId() {
        return merchantId;
    }

    public void setMerchantId(Long merchantId) {
        this.merchantId = merchantId;
    }

    public BigDecimal getAmount() {
        return amount;
    }

    public void setAmount(BigDecimal amount) {
        this.amount = amount;
    }

    public String getCurrency() {
        return currency;
    }

    public void setCurrency(String currency) {
        this.currency = currency;
    }

    public String getDeviceId() {
        return deviceId;
    }

    public void setDeviceId(String deviceId) {
        this.deviceId = deviceId;
    }

    public Instant getEventTime() {
        return eventTime;
    }

    public void setEventTime(Instant eventTime) {
        this.eventTime = eventTime;
    }

    public Instant getReceivedAt() {
        return receivedAt;
    }

    public void setReceivedAt(Instant receivedAt) {
        this.receivedAt = receivedAt;
    }

    public long getTxCountLast1Min() {
        return txCountLast1Min;
    }

    public void setTxCountLast1Min(long txCountLast1Min) {
        this.txCountLast1Min = txCountLast1Min;
    }

    public double getTxAmountLast1Hour() {
        return txAmountLast1Hour;
    }

    public void setTxAmountLast1Hour(double txAmountLast1Hour) {
        this.txAmountLast1Hour = txAmountLast1Hour;
    }

    public long getUniqueMerchantsLast24h() {
        return uniqueMerchantsLast24h;
    }

    public void setUniqueMerchantsLast24h(long uniqueMerchantsLast24h) {
        this.uniqueMerchantsLast24h = uniqueMerchantsLast24h;
    }

    public Instant getDeviceLastSeenAt() {
        return deviceLastSeenAt;
    }

    public void setDeviceLastSeenAt(Instant deviceLastSeenAt) {
        this.deviceLastSeenAt = deviceLastSeenAt;
    }

    public boolean isNewDevice() {
        return newDevice;
    }

    public void setNewDevice(boolean newDevice) {
        this.newDevice = newDevice;
    }

    public Instant getFeatureComputedAt() {
        return featureComputedAt;
    }

    public void setFeatureComputedAt(Instant featureComputedAt) {
        this.featureComputedAt = featureComputedAt;
    }
}
