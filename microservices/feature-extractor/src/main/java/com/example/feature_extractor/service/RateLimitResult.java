package com.example.feature_extractor.service;

public class RateLimitResult {

    private final boolean exceeded;
    private final double dailyUtilization;
    private final boolean dailyLimitExceeded;
    private final long senderTxCount24h;
    private final double senderTotalAmountUsd24h;
    private final long receiverInboundCount24h;
    private final long senderReceiverTxCount24h;

    public RateLimitResult(
            boolean exceeded,
            double dailyUtilization,
            boolean dailyLimitExceeded,
            long senderTxCount24h,
            double senderTotalAmountUsd24h,
            long receiverInboundCount24h,
            long senderReceiverTxCount24h
    ) {
        this.exceeded = exceeded;
        this.dailyUtilization = dailyUtilization;
        this.dailyLimitExceeded = dailyLimitExceeded;
        this.senderTxCount24h = senderTxCount24h;
        this.senderTotalAmountUsd24h = senderTotalAmountUsd24h;
        this.receiverInboundCount24h = receiverInboundCount24h;
        this.senderReceiverTxCount24h = senderReceiverTxCount24h;
    }

    public boolean isExceeded() {
        return exceeded;
    }

    public double getDailyUtilization() {
        return dailyUtilization;
    }

    public boolean isDailyLimitExceeded() {
        return dailyLimitExceeded;
    }

    public long getSenderTxCount24h() {
        return senderTxCount24h;
    }

    public double getSenderTotalAmountUsd24h() {
        return senderTotalAmountUsd24h;
    }

    public long getReceiverInboundCount24h() {
        return receiverInboundCount24h;
    }

    public long getSenderReceiverTxCount24h() {
        return senderReceiverTxCount24h;
    }
}
