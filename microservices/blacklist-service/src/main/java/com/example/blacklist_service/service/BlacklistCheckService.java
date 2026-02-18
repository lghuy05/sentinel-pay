package com.example.blacklist_service.service;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

import jakarta.annotation.PostConstruct;

import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import com.example.blacklist_service.entity.BlacklistEntry;
import com.example.blacklist_service.entity.BlacklistType;
import com.example.blacklist_service.event.BlacklistCheckEvent;
import com.example.blacklist_service.event.TransactionEnrichedEvent;
import com.example.blacklist_service.repository.BlacklistRepository;

@Service
public class BlacklistCheckService {

  private final BlacklistRepository blacklistRepository;
  private final StringRedisTemplate redisTemplate;

  public BlacklistCheckService(
      BlacklistRepository blacklistRepository,
      StringRedisTemplate redisTemplate) {
    this.blacklistRepository = blacklistRepository;
    this.redisTemplate = redisTemplate;
  }

  @PostConstruct
  public void warmCache() {
    List<BlacklistEntry> entries = blacklistRepository.findAllByActiveTrue();
    for (BlacklistEntry entry : entries) {
      String key = cacheKey(entry.getType(), entry.getValue());
      redisTemplate.opsForValue().set(key, "1");
    }
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
        matches.add("SENDER_USER_ID:" + event.getSenderUserId());
      }
    }

    if (event.getReceiverUserId() != null) {
      if (isBlacklisted(BlacklistType.USER_ID, String.valueOf(event.getReceiverUserId()))) {
        matches.add("RECEIVER_USER_ID:" + event.getReceiverUserId());
      }
    }

    if (event.getMerchantId() != null) {
      if (isBlacklisted(BlacklistType.MERCHANT_ID, String.valueOf(event.getMerchantId()))) {
        matches.add("MERCHANT_ID:" + event.getMerchantId());
      }
    }

    boolean hit = !matches.isEmpty();
    String reason = buildReason(matches);
    String decisionHint = hit ? "BLOCK" : null;

    return new BlacklistCheckEvent(
        event.getTransactionId(),
        hit,
        reason,
        decisionHint,
        event,
        Instant.now());
  }

  private boolean isBlacklisted(BlacklistType type, String value) {
    String key = cacheKey(type, value);
    String cached = redisTemplate.opsForValue().get(key);
    return "1".equals(cached);
  }

  private String cacheKey(BlacklistType type, String value) {
    String prefix;
    switch (type) {
      case DEVICE_ID:
        prefix = "blacklist:device:";
        break;
      case USER_ID:
        prefix = "blacklist:user:";
        break;
      case MERCHANT_ID:
        prefix = "blacklist:merchant:";
        break;
      default:
        prefix = "blacklist:account:";
        break;
    }
    return prefix + value;
  }

  private String buildReason(List<String> matches) {
    if (matches == null || matches.isEmpty()) {
      return null;
    }
    for (String entry : matches) {
      if (entry.startsWith("DEVICE_ID:")) {
        return "BLACKLISTED_DEVICE";
      }
      if (entry.startsWith("MERCHANT_ID:")) {
        return "BLACKLISTED_MERCHANT";
      }
      if (entry.startsWith("SENDER_USER_ID:")) {
        return "BLACKLISTED_SENDER";
      }
      if (entry.startsWith("RECEIVER_USER_ID:")) {
        return "BLACKLISTED_RECEIVER";
      }
    }
    return "BLACKLISTED_RECEIVER";
  }
}
