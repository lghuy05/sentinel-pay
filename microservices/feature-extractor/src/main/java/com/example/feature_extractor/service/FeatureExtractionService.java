package com.example.feature_extractor.service;

import java.time.Duration;
import java.time.Instant;

import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import com.example.feature_extractor.dto.AmountRiskTier;
import com.example.feature_extractor.event.TransactionEnrichedEvent;
import com.example.feature_extractor.event.TransactionReceivedEvent;

@Service
public class FeatureExtractionService {

    private static final Duration TX_COUNT_WINDOW = Duration.ofMinutes(1);
    private static final Duration TX_AMOUNT_WINDOW = Duration.ofHours(1);
    private static final Duration UNIQUE_MERCHANT_WINDOW = Duration.ofHours(24);
    private static final Duration DEVICE_LAST_SEEN_TTL = Duration.ofDays(7);

    private final StringRedisTemplate redisTemplate;
    private final RateLimitService rateLimitService;
    private final AccountClient accountClient;

    public FeatureExtractionService(
            StringRedisTemplate redisTemplate,
            RateLimitService rateLimitService,
            AccountClient accountClient
    ) {
        this.redisTemplate = redisTemplate;
        this.rateLimitService = rateLimitService;
        this.accountClient = accountClient;
    }

    public TransactionEnrichedEvent extract(TransactionReceivedEvent event) {
        AccountSnapshot sender = accountClient.fetchAccount(event.getSenderUserId());
        AccountSnapshot receiver = accountClient.fetchAccount(event.getReceiverUserId());

        String senderAccountCountry = normalizeCountry(sender.getAccountCountry());
        String receiverAccountCountry = normalizeCountry(receiver.getAccountCountry());
        boolean isCrossBorder = isCrossBorder(senderAccountCountry, receiverAccountCountry);
        boolean isOverseas = isCrossBorder;
        double amountUsdEquivalent = toUsdEquivalent(event.getAmount(), event.getCurrency());
        AmountRiskTier amountRiskTier = computeAmountRiskTier(amountUsdEquivalent);

        String senderKey = String.valueOf(event.getSenderUserId());
        long txCountLast1Min = incrementWithTtl(
                "feature:tx_count_1m:" + senderKey,
                1,
                TX_COUNT_WINDOW
        );
        double txAmountLast1Hour = incrementWithTtl(
                "feature:tx_amount_1h:" + senderKey,
                event.getAmount().doubleValue(),
                TX_AMOUNT_WINDOW
        );

        long uniqueMerchantsLast24h = 0;
        if (event.getMerchantId() != null) {
            String setKey = "feature:unique_merchants_24h:" + senderKey;
            redisTemplate.opsForSet().add(setKey, event.getMerchantId().toString());
            redisTemplate.expire(setKey, UNIQUE_MERCHANT_WINDOW);
            Long size = redisTemplate.opsForSet().size(setKey);
            uniqueMerchantsLast24h = size != null ? size : 0;
        }

        String deviceId = event.getDeviceId() != null ? event.getDeviceId() : "unknown";
        String deviceKey = "feature:device_last_seen:" + deviceId;
        String previous = redisTemplate.opsForValue().get(deviceKey);
        Instant previousSeen = previous != null ? Instant.ofEpochMilli(Long.parseLong(previous)) : null;
        boolean newDevice = previousSeen == null;
        redisTemplate.opsForValue().set(
                deviceKey,
                String.valueOf(Instant.now().toEpochMilli()),
                DEVICE_LAST_SEEN_TTL
        );

        boolean firstTimeContact = isFirstTimeContact(
                event.getSenderUserId(),
                event.getReceiverUserId(),
                event.getMerchantId()
        );

        RateLimitResult rateLimit = rateLimitService.evaluate(
                event.getSenderUserId(),
                event.getReceiverUserId(),
                event.getType(),
                event.getEventTime(),
                amountUsdEquivalent,
                isCrossBorder
        );

        String riskFlag = buildRiskFlag(rateLimit.isExceeded(), amountUsdEquivalent, event.getCurrency());

        return new TransactionEnrichedEvent(
                event.getTransactionId(),
                event.getType(),
                event.getSenderUserId(),
                event.getReceiverUserId(),
                event.getMerchantId(),
                event.getAmount(),
                event.getCurrency(),
                deviceId,
                event.getEventTime(),
                event.getReceivedAt(),
                senderAccountCountry,
                receiverAccountCountry,
                isCrossBorder,
                isOverseas,
                amountRiskTier,
                riskFlag,
                rateLimit.getDailyUtilization(),
                rateLimit.isDailyLimitExceeded(),
                firstTimeContact,
                amountUsdEquivalent,
                sender.getAccountAgeDays(),
                receiver.getAccountAgeDays(),
                rateLimit.getSenderTxCount24h(),
                rateLimit.getSenderTotalAmountUsd24h(),
                rateLimit.getReceiverInboundCount24h(),
                rateLimit.getSenderReceiverTxCount24h(),
                txCountLast1Min,
                txAmountLast1Hour,
                uniqueMerchantsLast24h,
                previousSeen,
                newDevice,
                Instant.now()
        );
    }

    private long incrementWithTtl(String key, long delta, Duration ttl) {
        Long updated = redisTemplate.opsForValue().increment(key, delta);
        ensureTtl(key, ttl);
        return updated != null ? updated : 0;
    }

    private double incrementWithTtl(String key, double delta, Duration ttl) {
        Double updated = redisTemplate.opsForValue().increment(key, delta);
        ensureTtl(key, ttl);
        return updated != null ? updated : 0.0;
    }

    private void ensureTtl(String key, Duration ttl) {
        Long expire = redisTemplate.getExpire(key);
        if (expire == null || expire < 0) {
            redisTemplate.expire(key, ttl);
        }
    }

    private boolean isCrossBorder(String senderCountry, String receiverCountry) {
        if (senderCountry == null || receiverCountry == null) {
            return false;
        }
        if ("UNKNOWN".equalsIgnoreCase(senderCountry) || "UNKNOWN".equalsIgnoreCase(receiverCountry)) {
            return false;
        }
        return !senderCountry.equalsIgnoreCase(receiverCountry);
    }

    private String normalizeCountry(String country) {
        if (country == null || country.isBlank()) {
            return "UNKNOWN";
        }
        return country.trim().toUpperCase();
    }

    private AmountRiskTier computeAmountRiskTier(double usdEquivalent) {
        if (usdEquivalent < 50) {
            return AmountRiskTier.LOW;
        }
        if (usdEquivalent < 300) {
            return AmountRiskTier.MEDIUM;
        }
        if (usdEquivalent < 1000) {
            return AmountRiskTier.HIGH;
        }
        return AmountRiskTier.CRITICAL;
    }

    private double toUsdEquivalent(java.math.BigDecimal amount, String currency) {
        if (amount == null || currency == null) {
            return 0.0;
        }
        double value = amount.doubleValue();
        switch (currency.toUpperCase()) {
            case "USD":
                return value;
            case "VND":
                return value / 25000.0;
            case "EUR":
                return value * 1.1;
            case "JPY":
                return value * 0.007;
            default:
                return value;
        }
    }

    private String buildRiskFlag(boolean rateLimitExceeded, double amountUsdEquivalent, String currency) {
        StringBuilder builder = new StringBuilder();
        if (rateLimitExceeded) {
            builder.append("RATE_LIMIT_EXCEEDED");
        }
        if (currency != null) {
            String normalized = currency.toUpperCase();
            if (!"USD".equals(normalized) && !"VND".equals(normalized)
                    && !"EUR".equals(normalized) && !"JPY".equals(normalized)) {
                if (builder.length() > 0) {
                    builder.append(",");
                }
                builder.append("UNKNOWN_CURRENCY");
            }
        }
        return builder.length() == 0 ? null : builder.toString();
    }

    private boolean isFirstTimeContact(Long senderUserId, Long receiverUserId, Long merchantId) {
        if (senderUserId == null) {
            return false;
        }
        String contactId = null;
        if (receiverUserId != null) {
            contactId = "user:" + receiverUserId;
        } else if (merchantId != null) {
            contactId = "merchant:" + merchantId;
        }
        if (contactId == null) {
            return false;
        }
        String key = "feature:first_contact:" + senderUserId;
        Long added = redisTemplate.opsForSet().add(key, contactId);
        redisTemplate.expire(key, Duration.ofDays(30));
        return added != null && added == 1L;
    }
}
