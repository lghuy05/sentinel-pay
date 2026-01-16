package com.example.feature_extractor.service;

public class AccountSnapshot {

    private String accountCountry = "UNKNOWN";
    private long accountAgeDays;
    private long txCount24h;
    private double totalAmountUsd24h;
    private long inboundCount24h;

    public String getAccountCountry() {
        return accountCountry;
    }

    public void setAccountCountry(String accountCountry) {
        this.accountCountry = accountCountry;
    }

    public long getAccountAgeDays() {
        return accountAgeDays;
    }

    public void setAccountAgeDays(long accountAgeDays) {
        this.accountAgeDays = accountAgeDays;
    }

    public long getTxCount24h() {
        return txCount24h;
    }

    public void setTxCount24h(long txCount24h) {
        this.txCount24h = txCount24h;
    }

    public double getTotalAmountUsd24h() {
        return totalAmountUsd24h;
    }

    public void setTotalAmountUsd24h(double totalAmountUsd24h) {
        this.totalAmountUsd24h = totalAmountUsd24h;
    }

    public long getInboundCount24h() {
        return inboundCount24h;
    }

    public void setInboundCount24h(long inboundCount24h) {
        this.inboundCount24h = inboundCount24h;
    }
}
