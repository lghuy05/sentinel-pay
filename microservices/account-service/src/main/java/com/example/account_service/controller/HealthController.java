package com.example.account_service.controller;

import java.util.Map;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class HealthController {

    @GetMapping("/health/account-service")
    public Map<String, String> health() {
        return Map.of("status", "UP");
    }
}
