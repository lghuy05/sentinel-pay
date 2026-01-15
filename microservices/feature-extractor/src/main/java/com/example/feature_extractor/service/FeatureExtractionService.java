package com.example.feature_extractor.service;

import java.time.Duration;
import java.time.Instant;

import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import com.example.feature_extractor.event.TransactionEnrichedEvent;
import com.example.feature_extractor.event.TransactionReceivedEvent;

@Service
public class FeatureExtractionService {

    private static final Duration TX_COUNT_WINDOW = Duration.ofMinutes(1);
    private static final Duration TX_AMOUNT_WINDOW = Duration.ofHours(1);
    private static final Duration UNIQUE_MERCHANT_WINDOW = Duration.ofHours(24);
    private static final Duration DEVICE_LAST_SEEN_TTL = Duration.ofDays(7);

    private final StringRedisTemplate redisTemplate;

    public FeatureExtractionService(StringRedisTemplate redisTemplate) {
        this.redisTemplate = redisTemplate;
    }

    public TransactionEnrichedEvent extract(TransactionReceivedEvent event) {
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

        return new TransactionEnrichedEvent(
                event.getTransactionId(),
                event.getType(),
                event.getSenderUserId(),
                event.getReceiverUserId(),
                event.getMerchantId(),
                event.getAmount(),
                event.getCurrency(),
                event.getIp(),
                deviceId,
                event.getEventTime(),
                event.getReceivedAt(),
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
}
