# 🚀 SentinelPay – Real-Time Fraud Detection System

**SentinelPay** is a distributed, event-driven fraud intelligence platform designed for Vietnamese mobile wallets.  
It combines **rule-based detection**, **machine learning**, and **real-time intelligence sharing** to prevent financial fraud at scale.

> 🎯 **Target Company:** MoMo (Vietnam FinTech)  
> 🕒 **Timeline:** Jan 13 – Feb 1 (3 weeks)  
> 📊 **Goal:** 60% production-ready system by Feb 1  
> 🟡 **Status:** In Development  

---

## 📌 Project Summary

- **Domain:** FinTech / Fraud Detection
- **Architecture:** Microservices + Event-Driven (Kafka)
- **Latency Target:** < 100ms per transaction
- **Focus:** Vietnamese fraud patterns (Tết scams, elderly scams, overseas fraud)
- **Deployment:** Kubernetes (local cluster)

---

## 🎯 Project Goals

### Technical
- Demonstrate **distributed systems** design
- Implement **real-time streaming with Kafka**
- Build **production-grade Spring Boot microservices**
- Deploy on **Kubernetes with monitoring**

### Business
- Solve **real Vietnamese fraud problems**
- Provide **explainable fraud decisions**
- Scale to **10,000+ transactions/min**

### Interview
- Strong **system design story**
- Clear **trade-offs & architecture reasoning**
- End-to-end **demoable project**

---

## 🧠 System Concept

**SentinelPay** processes transactions through multiple layers:

1. Transaction ingestion
2. Feature extraction
3. Parallel fraud checks:
   - Business rules
   - Blacklists
   - ML model
4. Fraud orchestration & decision
5. Alerts & monitoring

---

## 🔁 Event Flow (Kafka)

The services run concurrently and communicate through Kafka topics, so each stage can scale independently.

1. `transaction-ingestor` (REST) → `transactions.raw`
2. `feature-extractor` (Redis features) → `transactions.enriched`
3. Parallel checks (same input, different topics)
   - `rule-engine` → `fraud.rules`
   - `blacklist-service` → `fraud.blacklist`
   - `fraud-ml-service` → `fraud.ml`
4. `fraud-orchestrator` aggregates signals → `fraud.final`
5. `alert-service` persists decisions and triggers alerts

Note: `fraud-ml-service` is currently a stub/baseline scorer and should be replaced with a real trained model service.

---

## 📁 Project Structure

```text
sentinelpay/
├── microservices/                 # SPRING BOOT SERVICES (MAIN FOCUS)
│   ├── transaction-ingestor/      # Receive transactions (REST)
│   ├── feature-extractor/         # Real-time feature engineering
│   ├── rule-engine/               # Vietnamese fraud rules
│   ├── blacklist-service/         # Internal/external blacklist checks
│   ├── fraud-orchestrator/        # Final decision engine
│   └── alert-service/             # Fraud alerts & notifications
│
├── ml-services/                   # PYTHON (SIMPLE)
│   └── fraud-predictor/           # ML model serving (Flask)
│
├── infrastructure/                # Kafka, Redis, PostgreSQL
├── kubernetes/                    # K8s manifests
├── demo/                          # Demo scripts & Postman
└── docs/                          # Architecture & API docs

1. Client → POST /api/v1/transactions
   → transaction-ingestor

2. Validate & enrich
   → Kafka topic: transactions.raw

3. Feature extraction
   → Kafka topic: transactions.enriched

4. Parallel fraud checks
   ├── rule-engine → fraud.rules
   ├── blacklist-service → fraud.blacklist
   └── python ML service → fraud.ml

5. fraud-orchestrator
   → Combine all signals
   → Final decision (BLOCK / HOLD / ALLOW)
   → Kafka topic: fraud.final

6. alert-service
   → Send alerts
   → Update dashboard
