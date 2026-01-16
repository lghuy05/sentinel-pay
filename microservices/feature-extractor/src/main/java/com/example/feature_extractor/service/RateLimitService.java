package com.example.feature_extractor.service;

import java.time.Duration;
import java.time.Instant;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import com.example.feature_extractor.dto.TransactionType;

@Service
public class RateLimitService {

    private static final Logger log = LoggerFactory.getLogger(RateLimitService.class);

    private static final int HOURLY_TX_LIMIT = 10;
    private static final int DAILY_TX_LIMIT_P2P = 20;
    private static final int DAILY_TX_LIMIT_MERCHANT = 60;
    private static final int DAILY_SENDER_RECEIVER_LIMIT = 3;

    private static final long DAILY_DOMESTIC_LIMIT_VND = 50_000_000L;
    private static final long DAILY_CROSS_BORDER_LIMIT_VND = 10_000_000L;

    private final StringRedisTemplate redisTemplate;

    public RateLimitService(StringRedisTemplate redisTemplate) {
        this.redisTemplate = redisTemplate;
    }

    public RateLimitResult evaluate(
            Long senderUserId,
            Long receiverUserId,
            TransactionType type,
            Instant eventTime,
            double amountUsdEquivalent,
            boolean crossBorder
    ) {
        if (senderUserId == null || eventTime == null) {
            return new RateLimitResult(false, 0.0, false, 0L, 0.0, 0L, 0L);
        }

        try {
            String hourKey = buildHourKey(senderUserId, eventTime);
            String dayKey = buildDayKey(senderUserId, eventTime);

            long hourCount = incrementWithTtl(hourKey, 1, Duration.ofHours(2));
            long dayCount = incrementWithTtl(dayKey, 1, Duration.ofDays(2));

            long amountVnd = toVndFromUsd(amountUsdEquivalent);
            long dailyLimit = crossBorder ? DAILY_CROSS_BORDER_LIMIT_VND : DAILY_DOMESTIC_LIMIT_VND;
            String amountKey = buildAmountKey(senderUserId, eventTime, crossBorder);
            long totalAmount = incrementWithTtl(amountKey, amountVnd, Duration.ofDays(2));

            String amountUsdKey = buildUsdAmountKey(senderUserId, eventTime);
            double totalAmountUsd = incrementWithTtl(amountUsdKey, amountUsdEquivalent, Duration.ofDays(2));

            long senderReceiverCount = 0;
            if (receiverUserId != null) {
                String pairKey = buildPairKey(senderUserId, receiverUserId, eventTime);
                senderReceiverCount = incrementWithTtl(pairKey, 1, Duration.ofDays(2));
            }

            long receiverInboundCount = 0;
            if (receiverUserId != null) {
                String inboundKey = buildInboundKey(receiverUserId, eventTime);
                receiverInboundCount = incrementWithTtl(inboundKey, 1, Duration.ofDays(2));
            }

            int dailyLimitCount = type == TransactionType.MERCHANT_PAYMENT
                    ? DAILY_TX_LIMIT_MERCHANT
                    : DAILY_TX_LIMIT_P2P;
            boolean exceeded = hourCount > HOURLY_TX_LIMIT
                    || dayCount > dailyLimitCount
                    || totalAmount > dailyLimit
                    || senderReceiverCount > DAILY_SENDER_RECEIVER_LIMIT;
            double utilization = dailyLimit == 0 ? 0.0 : Math.min(1.5, (double) totalAmount / dailyLimit);
            boolean dailyLimitExceeded = totalAmount > dailyLimit;

            return new RateLimitResult(
                    exceeded,
                    utilization,
                    dailyLimitExceeded,
                    dayCount,
                    totalAmountUsd,
                    receiverInboundCount,
                    senderReceiverCount
            );
        } catch (Exception e) {
            log.warn("Rate limit evaluation failed", e);
            return new RateLimitResult(false, 0.0, false, 0L, 0.0, 0L, 0L);
        }
    }

    private long incrementWithTtl(String key, long delta, Duration ttl) {
        Long updated = redisTemplate.opsForValue().increment(key, delta);
        if (updated != null && updated == 1L) {
            redisTemplate.expire(key, ttl);
        }
        return updated != null ? updated : 0L;
    }

    private double incrementWithTtl(String key, double delta, Duration ttl) {
        Double updated = redisTemplate.opsForValue().increment(key, delta);
        if (updated != null && updated == delta) {
            redisTemplate.expire(key, ttl);
        }
        return updated != null ? updated : 0.0;
    }

    private String buildHourKey(Long senderUserId, Instant eventTime) {
        String hourBucket = DateTimeFormatter.ofPattern("yyyyMMddHH")
                .withZone(ZoneOffset.UTC)
                .format(eventTime);
        return "rate:tx:hour:" + senderUserId + ":" + hourBucket;
    }

    private String buildDayKey(Long senderUserId, Instant eventTime) {
        String dayBucket = DateTimeFormatter.ofPattern("yyyyMMdd")
                .withZone(ZoneOffset.UTC)
                .format(eventTime);
        return "rate:tx:day:" + senderUserId + ":" + dayBucket;
    }

    private String buildAmountKey(Long senderUserId, Instant eventTime, boolean crossBorder) {
        String dayBucket = DateTimeFormatter.ofPattern("yyyyMMdd")
                .withZone(ZoneOffset.UTC)
                .format(eventTime);
        String segment = crossBorder ? "cross" : "domestic";
        return "rate:amt:day:" + segment + ":" + senderUserId + ":" + dayBucket;
    }

    private String buildUsdAmountKey(Long senderUserId, Instant eventTime) {
        String dayBucket = DateTimeFormatter.ofPattern("yyyyMMdd")
                .withZone(ZoneOffset.UTC)
                .format(eventTime);
        return "rate:amt:usd:day:" + senderUserId + ":" + dayBucket;
    }

    private String buildInboundKey(Long receiverUserId, Instant eventTime) {
        String dayBucket = DateTimeFormatter.ofPattern("yyyyMMdd")
                .withZone(ZoneOffset.UTC)
                .format(eventTime);
        return "rate:inbound:day:" + receiverUserId + ":" + dayBucket;
    }

    private String buildPairKey(Long senderUserId, Long receiverUserId, Instant eventTime) {
        String dayBucket = DateTimeFormatter.ofPattern("yyyyMMdd")
                .withZone(ZoneOffset.UTC)
                .format(eventTime);
        return "rate:pair:day:" + senderUserId + ":" + receiverUserId + ":" + dayBucket;
    }

    private long toVndFromUsd(double amountUsdEquivalent) {
        return Math.round(amountUsdEquivalent * 25_000);
    }
}
