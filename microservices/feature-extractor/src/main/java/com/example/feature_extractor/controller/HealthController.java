package com.example.feature_extractor.controller;

import java.util.Map;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class HealthController {

    @GetMapping("/health/feature-extractor")
    public Map<String, String> health() {
        return Map.of("status", "UP");
    }
}
