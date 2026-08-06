# PaySim Full Benchmark - 2026-08-06

This note records the first full-dataset benchmark run for SentinelPay using the public PaySim dataset.

Dataset:

- source: https://www.kaggle.com/datasets/ealaxi/paysim1?resource=download
- local file: [microservices/ml-services/fraud-ml-service/training/paysim.csv](/home/huy/workspace/projects/sentinel-pay/microservices/ml-services/fraud-ml-service/training/paysim.csv)
- total rows: `6,362,620`

## Split strategy

The benchmark uses a time-based split from the PaySim `step` column:

- train: `70%` = `4,453,834`
- validation: `15%` = `954,393`
- test: `15%` = `954,393`

This is intentionally not a random split. SentinelPay is a streaming fraud system, so the model should train on earlier transactions and be evaluated on later transactions.

## Runtime-safe feature set

The full benchmark excludes `isFlaggedFraud`.

Reason:

- `isFlaggedFraud` is a dataset-side artifact
- it is not a trustworthy online inference field for SentinelPay
- keeping it would inflate the benchmark in a way the live system cannot reproduce

Feature set used in this run:

- `amount_normalized`
- `log_amount_normalized`
- `hour_of_day`
- `day_index`
- `is_transfer`
- `is_cash_out`
- `is_cash_in`
- `is_payment`
- `is_debit`
- `origin_is_merchant`
- `destination_is_merchant`
- `origin_balance_before_normalized`
- `destination_balance_before_normalized`
- `amount_over_origin_balance`
- `origin_balance_missing_or_zero`
- `destination_balance_missing_or_zero`

## Compared algorithms

The full benchmark compared the two strongest candidates from the earlier 500k-row screening run:

1. `random_forest`
2. `xgboost`

Artifacts:

- XGBoost metrics: [tmp/paysim-full-xgboost/paysim-metrics.json](/home/huy/workspace/projects/sentinel-pay/tmp/paysim-full-xgboost/paysim-metrics.json)
- Random Forest metrics: [tmp/paysim-full-random-forest/paysim-metrics.json](/home/huy/workspace/projects/sentinel-pay/tmp/paysim-full-random-forest/paysim-metrics.json)

## Results

### Random Forest

- threshold: `0.70`
- test accuracy: `0.9999926654952415`
- test precision: `1.0`
- test recall: `0.998253493013972`
- test F1: `0.9991259832688226`
- test ROC-AUC: `0.9999961602912539`
- test PR-AUC: `0.9998041643449176`
- confusion matrix:
  - TN: `950,385`
  - FP: `0`
  - FN: `7`
  - TP: `4,001`

Interpretation:

- SentinelPay caught `4,001 / 4,008` fraud transactions in the held-out future test window
- it produced `0` false positives in that run

### XGBoost

- threshold: `0.60`
- test accuracy: `0.999353515794856`
- test precision: `0.9896043892578689`
- test recall: `0.8550399201596807`
- test F1: `0.9174140008031053`
- test ROC-AUC: `0.9999595721408291`
- test PR-AUC: `0.990546155220461`
- confusion matrix:
  - TN: `950,349`
  - FP: `36`
  - FN: `581`
  - TP: `3,427`

Interpretation:

- SentinelPay caught `3,427 / 4,008` fraud transactions in the held-out future test window
- it missed `581` fraud transactions
- it produced `36` false positives

## Recommendation for this project

Current best benchmark model for SentinelPay on PaySim:

- `random_forest`

Reason:

- it has the highest fraud recall
- it has the highest F1
- it has the highest PR-AUC
- it produced `0` false positives in the held-out test window

## Why Random Forest is better here

Different algorithms win in different settings. For this project, Random Forest is better on the current PaySim benchmark because the fraud patterns are strongly separable with non-linear rules such as:

- transaction type combinations
- balance-before anomalies
- amount-to-balance ratio behavior
- merchant/account endpoint patterns
- time-step behavior

Random Forest handles these interactions well without needing linear assumptions.

Project statement:

> In SentinelPay, Random Forest outperformed XGBoost on the full PaySim benchmark because the dataset's fraud behavior is dominated by strong tree-friendly interaction patterns, especially transfer and cash-out structure plus account-balance anomalies. For this project, that made Random Forest the strongest detector on the held-out future transaction window.

## Why XGBoost can still be better in other scenarios

XGBoost is still a strong candidate and can be better when:

- the feature space is larger and noisier
- generalization matters more than fitting very sharp local patterns
- training cost and repeated retraining matter more
- missing values, regularization, and boosting behavior improve robustness

Project statement:

> XGBoost remains attractive for production-oriented fraud systems because it usually scales better for repeated retraining and often generalizes more smoothly on noisy tabular data, even when it is not the top scorer on a specific dataset.

## Operational note

During this benchmark, `xgboost` completed materially faster than `random_forest` in the same environment.

That means model choice should not be based on detection quality alone:

- if the goal is best PaySim benchmark score, choose `random_forest`
- if the goal is a stronger speed-to-retrain tradeoff, keep `xgboost` as the main alternative

## Live stack status

As of `2026-08-06`, the benchmark result has also been promoted into the running stack.

Verified:

- `sentinelpay-ui` production build succeeds
- Vite proxy still targets the expected gateway and service routes
- the `docker compose` stack starts successfully
- the gateway and ML endpoints respond live
- a transaction can be submitted through the gateway and return a final decision
- the running ML service reports `model_version = random_forest-paysim-benchmark-v1`

Important nuance:

- the live online feature mapper is currently a compatibility bridge from `transactions.enriched` to the benchmark feature space
- that path works end to end, but the canonical online/offline feature contract should still be formalized more explicitly

## Measured ingest performance on the benchmark host

This section documents the load test that was run against the live stack on `2026-08-06`.

Scope of the measurement:

- deployment mode: single-host Docker Compose
- measured endpoint: `POST /api/v1/transactions`
- path meaning: transaction-ingest acceptance through the gateway-backed API path
- this is not a multi-node Kubernetes benchmark
- this is not full `fraud.final` end-to-end decision latency

Host used:

- CPU: `Intel Core i7-9700K @ 3.60GHz`
- cores: `8`
- RAM: `15 GiB`
- container CPU/memory limits: not explicitly set in Compose

Method:

- tool: `k6`
- executor: constant-arrival-rate
- duration: `30s` per run
- endpoint base URL: `http://localhost:18082`
- unique transaction IDs and fresh accounts created in `setup()`

Measured results:

### 250 TPS target

- achieved request rate: `249.91 req/s`
- p95 request latency: `5.14 ms`

### 500 TPS target

- achieved request rate: `499.81 req/s`
- p95 request latency: `5.80 ms`

### 1000 TPS target

- achieved request rate: `999.20 req/s`
- p95 request latency: `7.07 ms`

### 1500 TPS target

- achieved request rate: `1499.04 req/s`
- p95 request latency: `19.97 ms`

### 2000 TPS target

- healthy sustained result was not maintained
- achieved request rate: about `1609.24 req/s`
- p95 request latency: `950.04 ms`
- dropped iterations: `4287`

Interpretation:

- on this host, the current stack safely sustained `1500 TPS` at about `20 ms p95` for the transaction-ingest API path
- on this host, the current stack safely sustained `1000+ TPS` at under `10 ms p95`
- the current healthy ceiling is materially below a stable `2000 TPS`

Observed bottlenecks during the overloaded run:

- `postgres`
- `transaction-ingestor`
- `api-gateway`

This means the next performance gains are more likely to come from hot-path write optimization, connection-pool tuning, and Postgres tuning than from broad tuning across every downstream service.

## Next steps

1. add exact benchmark timing capture into the evaluation scripts
2. formalize the canonical online/offline feature contract in code and docs
3. automate model promotion so the chosen artifact becomes the default without manual copy/rebuild steps
4. add a smoke assertion for expected `modelVersion` in the final decision response
