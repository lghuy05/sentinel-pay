# SentinelPay Fraud ML Service

Baseline ML microservice that publishes ML risk scores to Kafka. The current repo still contains local synthetic training helpers, but the target direction is to train against the public PaySim dataset rather than rely on self-created demo data.

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

## Dataset direction

Primary public dataset target:

- PaySim: https://www.kaggle.com/datasets/ealaxi/paysim1?resource=download

Read the repo note first:

- [docs/paysim-dataset-note.md](/home/huy/workspace/projects/sentinel-pay/docs/paysim-dataset-note.md)

Download PaySim into the repo:

```bash
python training/download_paysim.py
```

Important:

- PaySim is public and much stronger than the current repo-generated CSV
- PaySim is still simulator-generated, not a raw production ledger export
- for SentinelPay it should be treated as the main public benchmark dataset

## Local demo data

The scripts below are still useful for quick local experimentation, smoke work, and fallback demos.

### Generate data

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

## Benchmark evaluation

To evaluate how well the model detects fraud on an input dataset and emit predicted labels:

```bash
python training/evaluate_benchmark.py --data /path/to/dataset.csv --model logreg
python training/evaluate_benchmark.py --data training/paysim.csv --model xgboost --max-rows 1000000
```

This script:

- detects the supported dataset schema
- trains the selected benchmark model
- produces row-level predicted labels for the held-out test split
- reports recall, precision, ROC-AUC, PR-AUC, confusion matrix, and correct-detection percentage

To compare supported models side by side:

```bash
python training/compare_benchmarks.py --data training/paysim.csv --max-rows 1000000
```

Full-dataset benchmark note:

- [docs/paysim-full-benchmark-2026-08-06.md](/home/huy/workspace/projects/sentinel-pay/docs/paysim-full-benchmark-2026-08-06.md)

If the input file is PaySim:

- it uses a time-aware split from the `step` column

If the input file is the repo synthetic dataset:

- it falls back to a stratified random split

## Recommended modeling order

1. logistic regression baseline
2. tree ensembles: random forest, histogram gradient boosting, XGBoost
3. deeper models only if they materially improve recall and PR-AUC

The goal is to catch the most fraud with operationally manageable false positives. Do not optimize for plain accuracy.

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
