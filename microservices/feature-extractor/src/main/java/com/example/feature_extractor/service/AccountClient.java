package com.example.feature_extractor.service;

import java.time.Duration;
import java.time.Instant;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestClient;

@Service
public class AccountClient {

    private static final Logger log = LoggerFactory.getLogger(AccountClient.class);

    private final RestClient restClient;

    public AccountClient(@Value("${account.service.url:http://localhost:8087}") String baseUrl) {
        this.restClient = RestClient.builder()
                .baseUrl(baseUrl)
                .build();
    }

    public AccountSnapshot fetchAccount(Long userId) {
        AccountSnapshot snapshot = new AccountSnapshot();
        if (userId == null) {
            return snapshot;
        }
        try {
            AccountResponse response = restClient.get()
                    .uri("/api/v1/accounts/{userId}", userId)
                    .retrieve()
                    .onStatus(status -> status.isError(), (req, res) -> {
                        throw new IllegalStateException("Account service returned " + res.getStatusCode());
                    })
                    .body(AccountResponse.class);

            if (response == null) {
                return snapshot;
            }

            snapshot.setAccountCountry(response.getAccountCountry());
            snapshot.setHomeCurrency(response.getHomeCurrency());
            snapshot.setBalanceMinor(response.getBalanceMinor());
            snapshot.setAccountAgeDays(response.accountAgeDays());
            return snapshot;
        } catch (Exception e) {
            log.warn("Account lookup failed for userId={}", userId, e);
            return snapshot;
        }
    }

    static class AccountResponse {
        private Long userId;
        private String accountCountry;
        private String homeCurrency;
        private String createdAt;
        private long balanceMinor;

        public Long getUserId() {
            return userId;
        }

        public void setUserId(Long userId) {
            this.userId = userId;
        }

        public String getAccountCountry() {
            return accountCountry;
        }

        public void setAccountCountry(String accountCountry) {
            this.accountCountry = accountCountry;
        }

        public String getHomeCurrency() {
            return homeCurrency;
        }

        public void setHomeCurrency(String homeCurrency) {
            this.homeCurrency = homeCurrency;
        }

        public String getCreatedAt() {
            return createdAt;
        }

        public void setCreatedAt(String createdAt) {
            this.createdAt = createdAt;
        }

        public long getBalanceMinor() {
            return balanceMinor;
        }

        public void setBalanceMinor(long balanceMinor) {
            this.balanceMinor = balanceMinor;
        }

        public long accountAgeDays() {
            if (createdAt == null) {
                return 0L;
            }
            try {
                Instant created = Instant.parse(createdAt);
                long days = Duration.between(created, Instant.now()).toDays();
                return Math.max(days, 0L);
            } catch (Exception e) {
                return 0L;
            }
        }
    }
}
