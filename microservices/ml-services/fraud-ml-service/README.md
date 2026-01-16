# SentinelPay Fraud ML Service

Baseline ML microservice that trains a logistic regression fraud model from synthetic data and publishes ML risk scores to Kafka.

## Project structure

```
ml-service/
├── training/
│   ├── generate_data.py
│   ├── train_model.py
│   └── dataset.csv
├── inference/
│   ├── model.py
│   ├── kafka_consumer.py
│   ├── kafka_producer.py
│   └── main.py
├── model_artifacts/
│   └── fraud_model.joblib
├── requirements.txt
└── README.md
```

## Setup

```bash
python -m venv .venv
source .venv/bin/activate
pip install -r ml-service/requirements.txt
```

## Generate data

```bash
python ml-service/training/generate_data.py --rows 10000
```

To generate a realistic wallet dataset (CSV) for demos and analytics:

```bash
python ml-service/training/generate_wallet_dataset.py --rows 80000
```

## Train model

```bash
python ml-service/training/train_model.py
```

The model is saved to `ml-service/model_artifacts/fraud_model.joblib` along with the feature order and metadata.

## Run inference worker

From `ml-service/inference`:

```bash
python main.py
```

The worker consumes `transactions.enriched` events and publishes ML scores to `fraud.ml`.

## Kafka message contracts

Input (`transactions.enriched`) must include the training features:

```json
{
  "transactionId": "tx123",
  "amount": 1250000.0,
  "tx_count_1min": 2,
  "tx_amount_1hour": 2500000.0,
  "is_new_device": 0,
  "is_overseas": 1,
  "is_night": 0,
  "is_cross_border": 1,
  "amount_usd_equivalent": 50.0,
  "amount_risk_tier": "MEDIUM",
  "senderAccountAgeDays": 120,
  "receiverAccountAgeDays": 15,
  "senderTxCount24h": 4,
  "senderTotalAmountUsd24h": 120.0,
  "receiverInboundCount24h": 1,
  "is_first_time_receiver": 0
}
```

Output (`fraud.ml`):

```json
{
  "transactionId": "tx123",
  "mlScore": 0.87,
  "modelVersion": "logreg-synthetic-v1",
  "evaluatedAt": "2024-01-01T00:00:00+00:00"
}
```

## Configuration

Environment variables (optional):

- `KAFKA_BOOTSTRAP_SERVERS` (default: `localhost:19092`)
- `INPUT_TOPIC` (default: `transactions.enriched`)
- `OUTPUT_TOPIC` (default: `fraud.ml`)
- `KAFKA_GROUP_ID` (default: `fraud-ml-service`)
- `MODEL_PATH` (default: `../model_artifacts/fraud_model.joblib`)
- `MODEL_VERSION` (default: `logreg-synthetic-v1`)
- `HEALTH_PORT` (default: `8091`)

## Wallet behavior notes

US wallet transfers behave differently from bank wires: they skew small and frequent (food, splits, P2P),
with large transfers being rare and often risk-signaling. That is why large USD amounts and cross-border
flows are flagged at lower thresholds than bank systems. In VN, everyday spending is even smaller, and
very large wallet transfers are atypical, so they are treated as higher risk.
