package com.example.blacklist_service.service;

import java.time.Duration;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import com.example.blacklist_service.entity.BlacklistEntry;
import com.example.blacklist_service.entity.BlacklistType;
import com.example.blacklist_service.event.BlacklistCheckEvent;
import com.example.blacklist_service.event.TransactionEnrichedEvent;
import com.example.blacklist_service.repository.BlacklistRepository;

@Service
public class BlacklistCheckService {

    private static final Duration CACHE_TTL = Duration.ofMinutes(10);

    private final BlacklistRepository blacklistRepository;
    private final StringRedisTemplate redisTemplate;

    public BlacklistCheckService(
            BlacklistRepository blacklistRepository,
            StringRedisTemplate redisTemplate
    ) {
        this.blacklistRepository = blacklistRepository;
        this.redisTemplate = redisTemplate;
    }

    public BlacklistCheckEvent check(TransactionEnrichedEvent event) {
        List<String> matches = new ArrayList<>();

        if (event.getDeviceId() != null) {
            if (isBlacklisted(BlacklistType.DEVICE_ID, event.getDeviceId())) {
                matches.add("DEVICE_ID:" + event.getDeviceId());
            }
        }

        if (event.getSenderUserId() != null) {
            if (isBlacklisted(BlacklistType.USER_ID, String.valueOf(event.getSenderUserId()))) {
                matches.add("USER_ID:" + event.getSenderUserId());
            }
        }

        if (event.getMerchantId() != null) {
            if (isBlacklisted(BlacklistType.MERCHANT_ID, String.valueOf(event.getMerchantId()))) {
                matches.add("MERCHANT_ID:" + event.getMerchantId());
            }
        }

        double score = matches.isEmpty() ? 0.0 : 1.0;

        return new BlacklistCheckEvent(
                event.getTransactionId(),
                score,
                matches,
                Instant.now()
        );
    }

    private boolean isBlacklisted(BlacklistType type, String value) {
        String key = cacheKey(type, value);
        String cached = redisTemplate.opsForValue().get(key);
        if ("1".equals(cached)) {
            return true;
        }
        if ("0".equals(cached)) {
            return false;
        }

        Optional<BlacklistEntry> entry =
                blacklistRepository.findByTypeAndValueAndActiveTrue(type, value);
        boolean hit = entry.isPresent();
        redisTemplate.opsForValue().set(key, hit ? "1" : "0", CACHE_TTL);
        return hit;
    }

    private String cacheKey(BlacklistType type, String value) {
        return "blacklist:" + type.name() + ":" + value;
    }
}
