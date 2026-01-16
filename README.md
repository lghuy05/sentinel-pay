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
- **Focus:** Vietnamese fraud patterns (Tết scams, elderly scams, cross-border fraud)
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

Note: `fraud-ml-service` is a baseline model and should be retrained regularly with updated synthetic data.

---

## 📁 Project Structure

```text
sentinelpay/
├── microservices/                 # SPRING BOOT SERVICES (MAIN FOCUS)
│   ├── account-service/           # Account metadata for testing + enrichment
│   ├── transaction-ingestor/      # Receive transactions (REST)
│   ├── feature-extractor/         # Real-time feature engineering
│   ├── rule-engine/               # Vietnamese fraud rules
│   ├── blacklist-service/         # Internal/external blacklist checks
│   ├── fraud-orchestrator/        # Final decision engine
│   └── alert-service/             # Fraud alerts & notifications
│
├── microservices/ml-services/     # PYTHON (SIMPLE)
│   └── fraud-ml-service/          # ML model worker (Kafka)
│
├── infrastructure/                # Kafka, Redis, PostgreSQL
├── kubernetes/                    # K8s manifests
├── demo/                          # Demo scripts & Postman
├── sentinelpay-ui/                # Vue 3 Ops Console
└── docs/                          # Architecture & API docs
```

---

## ✅ Quick Start (Local)

Start infrastructure + backend services:

```bash
./scripts/start-services.sh
```

Start the Ops Console:

```bash
cd sentinelpay-ui
npm install
npm run dev
```

ML training (optional but recommended):

```bash
cd microservices/ml-services/fraud-ml-service
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python training/generate_data.py --rows 10000
python training/train_model.py
```

---

## 🧪 Demo Flow (Ops Console)

1. Open the Ops Console at `http://localhost:5173`.
2. Go to **Accounts** and create:
   - US sender (1y old, $50k)
   - VN receiver (2d old, 0 VND)
3. Go to **Simulate Transaction** and send:
   - 2,000,000 USD → Expect **BLOCK** (hard stop)
   - 4,000 USD cross-border → Expect **REVIEW**
   - 20 USD domestic → Expect **ALLOW**
4. Inspect results in **Fraud Decisions** and **Dashboard**.

---

## 🧩 Ops Console Notes

- The UI pulls decisions from `alert-service` via `/api/v1/decisions`.
- Health checks hit `/health/<service>` endpoints across services.
- Transactions are submitted to `transaction-ingestor` and decisions are fetched asynchronously.
