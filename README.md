# SentinelPay

**SentinelPay** is a real-time fraud detection system for digital wallets. It was built to show how a modern payment platform can catch risky transactions fast, explain why a decision was made, and still give operators a clean way to review cases live during a demo.

## The Problem

Digital wallet fraud is difficult because bad transactions do not all look the same.

- Some attacks are obvious, like blacklisted users or devices.
- Some are behavioral, like sudden high-value transfers at unusual hours.
- Some look normal to simple rule engines but become suspicious when account age, device history, velocity, and cross-border context are combined.

A basic app can store transactions. A real fraud platform has to **detect, score, explain, and react in real time**.

That is the gap SentinelPay is designed to solve.

## The Solution

SentinelPay processes each transaction as an event and sends it through a fraud pipeline:

1. ingest the transaction
2. enrich it with context
3. check blacklist risk
4. apply fraud rules
5. score it with ML
6. combine signals into one explainable final decision
7. surface the result in an ops dashboard

This gives the system two important properties:

- **speed**, because each stage is event-driven and isolated
- **explainability**, because the final outcome is based on visible signals instead of a black-box output alone

## Demo Preview

![img-alt](https://github.com/lghuy05/sentinel-pay/blob/f6b7f093feb3a755b93a188096adca73a571b315/Untitled-2025-12-27-1630.png)
![img-alt](https://github.com/lghuy05/sentinel-pay/blob/f6b7f093feb3a755b93a188096adca73a571b315/1.png)
![img-alt](https://github.com/lghuy05/sentinel-pay/blob/f6b7f093feb3a755b93a188096adca73a571b315/2.png)
![img-alt](https://github.com/lghuy05/sentinel-pay/blob/f6b7f093feb3a755b93a188096adca73a571b315/3.png)
![img-alt](https://github.com/lghuy05/sentinel-pay/blob/f6b7f093feb3a755b93a188096adca73a571b315/4.png)
![img-alt](https://github.com/lghuy05/sentinel-pay/blob/f6b7f093feb3a755b93a188096adca73a571b315/5.png)
![img-alt](https://github.com/lghuy05/sentinel-pay/blob/f6b7f093feb3a755b93a188096adca73a571b315/6.png)

## How It Works

### 1. Transaction ingestion

A wallet transfer enters the system and is accepted immediately for processing.

### 2. Feature extraction

The system adds context that raw transactions do not carry by themselves, such as:

- account age
- transaction frequency
- recent amount totals
- device familiarity
- cross-border behavior
- time-of-day patterns

### 3. Blacklist screening

Known bad entities are checked early so obviously risky traffic can be escalated fast.

### 4. Rule evaluation

Deterministic fraud rules score suspicious patterns such as:

- large or unusual transfers
- risky device changes
- suspicious timing
- new-account behavior
- high-risk country combinations

### 5. ML scoring

A machine learning model adds a probability-based fraud score for patterns that are harder to capture with hand-written rules alone.

### 6. Final orchestration

The orchestrator combines blacklist, rules, and ML into one final decision such as:

- `ALLOW`
- `HOLD`
- `BLOCK`

### 7. Analyst feedback

Borderline cases can be reviewed in the dashboard and labeled later, creating a feedback loop for model improvement.

## Decision Formula

SentinelPay is designed around a simple idea:

```text
Final Risk = blacklist signal + rule score + ML score + behavioral context
```

Conceptually:

```text
if blacklist hit -> BLOCK
else if combined risk is very high -> BLOCK
else if combined risk is uncertain -> HOLD
else -> ALLOW
```

This is what makes the project demo-friendly: the decision is not just a number. It can be explained as a combination of explicit rules and learned risk.

## What Makes This Project Special

### Explainable fraud decisions

Many student ML projects stop at prediction. SentinelPay makes the result readable for operators by showing the rule score, ML score, and final reason together.

### Hybrid detection

It does not rely only on rules or only on ML. It uses both.

- rules catch explicit patterns
- ML catches subtle behavioral patterns
- orchestration decides the final outcome

### Event-driven architecture

The pipeline is built around asynchronous processing, which makes the project closer to a real financial system than a typical monolith demo.

### Full demo loop

The project is not only backend logic. It includes:

- account setup
- transaction simulation
- fraud decision review
- analyst feedback
- system health visibility

That makes it easy to present live to recruiters, engineers, or judges.

### Real operations feel

The UI behaves like a small fraud operations console, not just a generic admin panel.

## System Flow

```text
Transaction
   ->
Ingestion
   ->
Feature Enrichment
   ->
Blacklist Check
   ->
Rule Evaluation
   ->
ML Scoring
   ->
Fraud Orchestration
   ->
Final Decision
   ->
Dashboard / Alerts / Analyst Feedback
```

## Runtime

The active backend stack is Go for the platform services and Python for the ML service. The legacy Java services have been removed from the repository after the Go migration was merged and verified.

## Current Project Status

As of August 6, 2026, SentinelPay has reached these milestones:

- the platform services run as Go microservices
- the ML service runs separately in Python
- the frontend still builds and targets the current gateway routes
- the live stack starts with Docker Compose
- a transaction can flow end to end from ingestion to final fraud decision
- the ML service has been benchmarked on the full public PaySim dataset
- the live ML service has been promoted from the old synthetic logistic-regression model to the PaySim-backed Random Forest model

## What The Project Has Achieved

### 1. End-to-end fraud pipeline is live

SentinelPay is no longer just an architecture plan or an offline model demo.

It now runs as a working multi-service fraud pipeline:

- transaction ingestion
- feature enrichment
- blacklist check
- rule scoring
- ML scoring
- fraud orchestration
- final decision exposure through HTTP

The live stack was verified by sending transactions through the gateway and reading final decisions back from the decisions API.

### 2. Go microservice migration is complete

The platform services that previously depended on Java have been replaced in Go while keeping the project as a microservice system rather than collapsing it into a monolith.

That means the project now demonstrates:

- service boundaries
- HTTP contracts
- Kafka event flow
- Redis usage
- Postgres-backed persistence
- Docker Compose orchestration

without carrying the old Java runtime in the repository.

### 3. The project uses a real public fraud dataset

The ML evaluation path is no longer limited to self-generated synthetic CSV data.

SentinelPay now uses the public PaySim dataset:

- dataset size: `6,362,620` rows
- source: https://www.kaggle.com/datasets/ealaxi/paysim1?resource=download

This gives the project a stronger interview and demo story because the model evaluation is based on a recognized public fraud dataset instead of only handcrafted samples.

### 4. Full-dataset model benchmarking is documented

SentinelPay now has a documented benchmark process and stored benchmark artifacts for full-dataset evaluation.

Reference:

- [docs/paysim-full-benchmark-2026-08-06.md](/home/huy/workspace/projects/sentinel-pay/docs/paysim-full-benchmark-2026-08-06.md)

Benchmark setup:

- time-based split
- train: `70%`
- validation: `15%`
- test: `15%`

This matters because the project evaluates fraud detection more like a real streaming system: training on earlier transactions and testing on later ones.

### 5. Best current model is Random Forest on PaySim

The strongest current full-dataset benchmark result is:

- model: `random_forest-paysim-benchmark-v1`

Held-out future test results:

- precision: `1.0`
- recall: `0.998253493013972`
- F1: `0.9991259832688226`
- PR-AUC: `0.9998041643449176`

Compared full-dataset alternative:

- `xgboost`
- precision: `0.9896043892578689`
- recall: `0.8550399201596807`
- F1: `0.9174140008031053`
- PR-AUC: `0.990546155220461`

For this project, Random Forest currently performs better because the PaySim fraud patterns are strongly separable with tree-friendly interactions such as:

- transfer and payment type structure
- balance anomalies
- amount-to-balance ratios
- transaction timing behavior

### 6. The stack has measured single-host throughput

SentinelPay has also been load-tested on the live Docker Compose stack on the benchmark host.

Measured transaction-ingest API result:

- sustained `1500 TPS`
- about `20 ms p95`
- single `8-core / 15 GB` host

Important scope note:

- this number applies to the gateway-backed transaction-ingest API path
- it is not a multi-node Kubernetes benchmark
- it is not the same as full end-to-end `fraud.final` latency

### 7. The live stack is serving the new model

The running ML service has already been switched to the new benchmarked model, not just evaluated offline.

Live model version:

- `random_forest-paysim-benchmark-v1`

That means the current running system can produce live fraud decisions using the new PaySim-backed model through the same API path the frontend uses.

## What To Say In A Demo Or Interview

Short version:

> SentinelPay is a real-time fraud detection platform for digital wallets built with Go microservices, Kafka, Redis, PostgreSQL, a Python ML service, and a live ops dashboard. It processes transactions end to end, explains decisions with rules plus ML, and has been benchmarked on the 6.36 million-row PaySim public fraud dataset. The current deployed model is a Random Forest that outperformed XGBoost on the held-out future transaction window.

## Important Scope Note

The project is working end to end, but the ML feature contract is still in a transitional state:

- the live service can now map `transactions.enriched` into the PaySim-style benchmark feature space
- that mapper is a compatibility bridge
- the next cleanup step is to formalize that canonical online/offline feature contract in code and docs
