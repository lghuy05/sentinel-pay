package com.example.blacklist_service.event;

import java.math.BigDecimal;
import java.time.Instant;

import com.example.blacklist_service.dto.TransactionType;
import com.fasterxml.jackson.annotation.JsonAlias;
import com.fasterxml.jackson.annotation.JsonProperty;

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
    private String senderAccountCountry;
    private String receiverAccountCountry;
    private String senderHomeCurrency;
    private long senderBalanceMinor;
    @JsonProperty("is_cross_border")
    private boolean crossBorder;
    @JsonProperty("is_overseas")
    private boolean overseas;
    @JsonProperty("amount_risk_tier")
    private String amountRiskTier;
    private String riskFlag;
    private double dailyAmountUtilization;
    private boolean dailyLimitExceeded;
    @JsonProperty("is_first_time_receiver")
    private boolean firstTimeContact;
    @JsonProperty("amount_usd_equivalent")
    private double amountUsdEquivalent;
    private long senderAccountAgeDays;
    private long receiverAccountAgeDays;
    private long senderTxCount24h;
    private double senderTotalAmountUsd24h;
    private long receiverInboundCount24h;
    @JsonProperty("sender_receiver_tx_count_24h")
    private long senderReceiverTxCount24h;
    @JsonProperty("tx_count_1m")
    @JsonAlias("tx_count_1min")
    private long txCountLast1Min;
    @JsonProperty("tx_count_1h")
    private long txCountLast1Hour;
    @JsonProperty("tx_amount_1hour")
    private double txAmountLast1Hour;
    @JsonProperty("amount_1h")
    private double amountLast1Hour;
    private long uniqueMerchantsLast24h;
    @JsonProperty("last_tx_time")
    private Instant lastTxTime;
    @JsonProperty("time_since_last_tx")
    private Long timeSinceLastTxSeconds;
    private Instant deviceLastSeenAt;
    @JsonProperty("is_new_device")
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

    public String getSenderAccountCountry() {
        return senderAccountCountry;
    }

    public void setSenderAccountCountry(String senderAccountCountry) {
        this.senderAccountCountry = senderAccountCountry;
    }

    public String getReceiverAccountCountry() {
        return receiverAccountCountry;
    }

    public void setReceiverAccountCountry(String receiverAccountCountry) {
        this.receiverAccountCountry = receiverAccountCountry;
    }

    public String getSenderHomeCurrency() {
        return senderHomeCurrency;
    }

    public void setSenderHomeCurrency(String senderHomeCurrency) {
        this.senderHomeCurrency = senderHomeCurrency;
    }

    public long getSenderBalanceMinor() {
        return senderBalanceMinor;
    }

    public void setSenderBalanceMinor(long senderBalanceMinor) {
        this.senderBalanceMinor = senderBalanceMinor;
    }

    public boolean isCrossBorder() {
        return crossBorder;
    }

    public void setCrossBorder(boolean crossBorder) {
        this.crossBorder = crossBorder;
    }

    public boolean isOverseas() {
        return overseas;
    }

    public void setOverseas(boolean overseas) {
        this.overseas = overseas;
    }

    public String getAmountRiskTier() {
        return amountRiskTier;
    }

    public void setAmountRiskTier(String amountRiskTier) {
        this.amountRiskTier = amountRiskTier;
    }

    public String getRiskFlag() {
        return riskFlag;
    }

    public void setRiskFlag(String riskFlag) {
        this.riskFlag = riskFlag;
    }

    public double getDailyAmountUtilization() {
        return dailyAmountUtilization;
    }

    public void setDailyAmountUtilization(double dailyAmountUtilization) {
        this.dailyAmountUtilization = dailyAmountUtilization;
    }

    public boolean isDailyLimitExceeded() {
        return dailyLimitExceeded;
    }

    public void setDailyLimitExceeded(boolean dailyLimitExceeded) {
        this.dailyLimitExceeded = dailyLimitExceeded;
    }

    public boolean isFirstTimeContact() {
        return firstTimeContact;
    }

    public void setFirstTimeContact(boolean firstTimeContact) {
        this.firstTimeContact = firstTimeContact;
    }

    public double getAmountUsdEquivalent() {
        return amountUsdEquivalent;
    }

    public void setAmountUsdEquivalent(double amountUsdEquivalent) {
        this.amountUsdEquivalent = amountUsdEquivalent;
    }

    public long getSenderAccountAgeDays() {
        return senderAccountAgeDays;
    }

    public void setSenderAccountAgeDays(long senderAccountAgeDays) {
        this.senderAccountAgeDays = senderAccountAgeDays;
    }

    public long getReceiverAccountAgeDays() {
        return receiverAccountAgeDays;
    }

    public void setReceiverAccountAgeDays(long receiverAccountAgeDays) {
        this.receiverAccountAgeDays = receiverAccountAgeDays;
    }

    public long getSenderTxCount24h() {
        return senderTxCount24h;
    }

    public void setSenderTxCount24h(long senderTxCount24h) {
        this.senderTxCount24h = senderTxCount24h;
    }

    public double getSenderTotalAmountUsd24h() {
        return senderTotalAmountUsd24h;
    }

    public void setSenderTotalAmountUsd24h(double senderTotalAmountUsd24h) {
        this.senderTotalAmountUsd24h = senderTotalAmountUsd24h;
    }

    public long getReceiverInboundCount24h() {
        return receiverInboundCount24h;
    }

    public void setReceiverInboundCount24h(long receiverInboundCount24h) {
        this.receiverInboundCount24h = receiverInboundCount24h;
    }

    public long getSenderReceiverTxCount24h() {
        return senderReceiverTxCount24h;
    }

    public void setSenderReceiverTxCount24h(long senderReceiverTxCount24h) {
        this.senderReceiverTxCount24h = senderReceiverTxCount24h;
    }

    public long getTxCountLast1Min() {
        return txCountLast1Min;
    }

    public void setTxCountLast1Min(long txCountLast1Min) {
        this.txCountLast1Min = txCountLast1Min;
    }

    public long getTxCountLast1Hour() {
        return txCountLast1Hour;
    }

    public void setTxCountLast1Hour(long txCountLast1Hour) {
        this.txCountLast1Hour = txCountLast1Hour;
    }

    public double getTxAmountLast1Hour() {
        return txAmountLast1Hour;
    }

    public void setTxAmountLast1Hour(double txAmountLast1Hour) {
        this.txAmountLast1Hour = txAmountLast1Hour;
    }

    public double getAmountLast1Hour() {
        return amountLast1Hour;
    }

    public void setAmountLast1Hour(double amountLast1Hour) {
        this.amountLast1Hour = amountLast1Hour;
    }

    public long getUniqueMerchantsLast24h() {
        return uniqueMerchantsLast24h;
    }

    public void setUniqueMerchantsLast24h(long uniqueMerchantsLast24h) {
        this.uniqueMerchantsLast24h = uniqueMerchantsLast24h;
    }

    public Instant getLastTxTime() {
        return lastTxTime;
    }

    public void setLastTxTime(Instant lastTxTime) {
        this.lastTxTime = lastTxTime;
    }

    public Long getTimeSinceLastTxSeconds() {
        return timeSinceLastTxSeconds;
    }

    public void setTimeSinceLastTxSeconds(Long timeSinceLastTxSeconds) {
        this.timeSinceLastTxSeconds = timeSinceLastTxSeconds;
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
