package com.example.fraud_orchestrator.controller;

import java.time.Duration;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;

import org.apache.kafka.clients.admin.AdminClient;
import org.apache.kafka.clients.admin.ListTopicsResult;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.data.redis.connection.RedisConnectionFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.kafka.core.KafkaAdmin;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.client.RestClient;

@RestController
@RequestMapping("/system")
public class SystemController {

    private final KafkaAdmin kafkaAdmin;
    private final RedisConnectionFactory redisConnectionFactory;
    private final RestClient restClient;
    private final Map<String, String> serviceUrls;

    public SystemController(
            KafkaAdmin kafkaAdmin,
            RedisConnectionFactory redisConnectionFactory,
            @Value("${services.ingestor.url:http://localhost:8081/health/transaction-ingestor}") String ingestorUrl,
            @Value("${services.extractor.url:http://localhost:8082/health/feature-extractor}") String extractorUrl,
            @Value("${services.blacklist.url:http://localhost:8084/health/blacklist-service}") String blacklistUrl,
            @Value("${services.rule.url:http://localhost:8083/health/rule-engine}") String ruleUrl,
            @Value("${services.ml.url:http://localhost:8091/health/ml-service}") String mlUrl,
            @Value("${services.orchestrator.url:http://localhost:8085/actuator/health}") String orchestratorUrl
    ) {
        this.kafkaAdmin = kafkaAdmin;
        this.redisConnectionFactory = redisConnectionFactory;
        this.restClient = RestClient.builder().build();
        this.serviceUrls = Map.of(
                "ingestor", ingestorUrl,
                "extractor", extractorUrl,
                "blacklist", blacklistUrl,
                "rule-engine", ruleUrl,
                "ml-service", mlUrl,
                "orchestrator", orchestratorUrl
        );
    }

    @GetMapping("/kafka")
    public ResponseEntity<Map<String, Object>> kafka() {
        Map<String, Object> payload = new HashMap<>();
        try (AdminClient admin = AdminClient.create(kafkaAdmin.getConfigurationProperties())) {
            ListTopicsResult topicsResult = admin.listTopics();
            payload.put("status", "UP");
            payload.put("topics", topicsResult.names().get(3, TimeUnit.SECONDS));
        } catch (Exception e) {
            payload.put("status", "DOWN");
            payload.put("error", e.getMessage());
        }
        return ResponseEntity.ok(payload);
    }

    @GetMapping("/redis")
    public ResponseEntity<Map<String, Object>> redis() {
        Map<String, Object> payload = new HashMap<>();
        try {
            var connection = redisConnectionFactory.getConnection();
            var info = connection.info();
            payload.put("status", "UP");
            payload.put("info", info);
            connection.close();
        } catch (Exception e) {
            payload.put("status", "DOWN");
            payload.put("error", e.getMessage());
        }
        return ResponseEntity.ok(payload);
    }

    @GetMapping("/services")
    public ResponseEntity<Map<String, Object>> services() {
        Map<String, Object> payload = new HashMap<>();
        for (Map.Entry<String, String> entry : serviceUrls.entrySet()) {
            payload.put(entry.getKey(), probe(entry.getValue()));
        }
        return ResponseEntity.ok(payload);
    }

    private Map<String, Object> probe(String url) {
        Map<String, Object> status = new HashMap<>();
        try {
            Map response = restClient.get()
                    .uri(url)
                    .retrieve()
                    .body(Map.class);
            status.put("status", "UP");
            status.put("response", response);
        } catch (Exception e) {
            status.put("status", "DOWN");
            status.put("error", e.getMessage());
        }
        return status;
    }
}
