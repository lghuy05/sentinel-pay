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

## Why This Is Strong For Intern Applications

This project shows the kind of engineering breadth that stands out:

- backend architecture
- distributed event flow
- fraud/risk product thinking
- applied ML in a real pipeline
- a polished frontend for live demos
- human-in-the-loop review instead of blind automation

It reads like a real product, not a tutorial clone.

## Elevator Pitch

SentinelPay is a real-time fraud detection platform for digital wallets that combines event-driven microservices, fraud rules, ML scoring, and an analyst dashboard to turn raw transactions into explainable fraud decisions.

## Runtime

The active backend stack is Go for the platform services and Python for the ML service. The legacy Java services have been removed from the repository after the Go migration was merged and verified.
