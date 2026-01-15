package com.example.feature_extractor.event;

import java.math.BigDecimal;
import java.time.Instant;

import com.example.feature_extractor.dto.TransactionType;

public class TransactionEnrichedEvent {

    private String transactionId;
    private TransactionType type;
    private Long senderUserId;
    private Long receiverUserId;
    private Long merchantId;
    private BigDecimal amount;
    private String currency;
    private String ip;
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

    public TransactionEnrichedEvent(
            String transactionId,
            TransactionType type,
            Long senderUserId,
            Long receiverUserId,
            Long merchantId,
            BigDecimal amount,
            String currency,
            String ip,
            String deviceId,
            Instant eventTime,
            Instant receivedAt,
            long txCountLast1Min,
            double txAmountLast1Hour,
            long uniqueMerchantsLast24h,
            Instant deviceLastSeenAt,
            boolean newDevice,
            Instant featureComputedAt
    ) {
        this.transactionId = transactionId;
        this.type = type;
        this.senderUserId = senderUserId;
        this.receiverUserId = receiverUserId;
        this.merchantId = merchantId;
        this.amount = amount;
        this.currency = currency;
        this.ip = ip;
        this.deviceId = deviceId;
        this.eventTime = eventTime;
        this.receivedAt = receivedAt;
        this.txCountLast1Min = txCountLast1Min;
        this.txAmountLast1Hour = txAmountLast1Hour;
        this.uniqueMerchantsLast24h = uniqueMerchantsLast24h;
        this.deviceLastSeenAt = deviceLastSeenAt;
        this.newDevice = newDevice;
        this.featureComputedAt = featureComputedAt;
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

    public String getIp() {
        return ip;
    }

    public void setIp(String ip) {
        this.ip = ip;
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
