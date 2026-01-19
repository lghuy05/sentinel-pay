package com.example.fraud_orchestrator.controller;

import java.util.Map;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.client.RestClient;

@RestController
@RequestMapping("/ml")
public class MlController {

    private final RestClient restClient;

    public MlController(@Value("${ml.service.url:http://localhost:8091}") String baseUrl) {
        this.restClient = RestClient.builder().baseUrl(baseUrl).build();
    }

    @GetMapping("/status")
    public ResponseEntity<Map> status() {
        Map body = restClient.get()
                .uri("/ml/status")
                .retrieve()
                .body(Map.class);
        return ResponseEntity.ok(body);
    }

    @PostMapping("/retrain")
    public ResponseEntity<Map> retrain() {
        Map body = restClient.post()
                .uri("/ml/retrain")
                .retrieve()
                .body(Map.class);
        return ResponseEntity.ok(body);
    }

    @PostMapping("/reload")
    public ResponseEntity<Map> reload() {
        Map body = restClient.post()
                .uri("/ml/reload")
                .retrieve()
                .body(Map.class);
        return ResponseEntity.ok(body);
    }
}
