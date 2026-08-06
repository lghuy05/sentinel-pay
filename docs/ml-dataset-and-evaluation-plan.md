# ML Dataset And Evaluation Plan

## Goal

Move SentinelPay fraud modeling from repo-generated demo data to a stronger public benchmark dataset while keeping the microservice product model stable.

Primary dataset target:

- PaySim: https://www.kaggle.com/datasets/ealaxi/paysim1?resource=download

Supporting note:

- [paysim-dataset-note.md](/home/huy/workspace/projects/sentinel-pay/docs/paysim-dataset-note.md)

## Key Architecture Decision

Do not reshape the full product schema around one Kaggle CSV.

Instead, define three layers:

1. raw dataset schema
2. canonical fraud feature schema
3. online inference schema

That allows:

- many datasets in the future
- stable frontend and service contracts
- offline training on public datasets
- online scoring on live SentinelPay events

## Data Model Plan

### Layer 1: Raw dataset schemas

Add dataset-specific loaders and validators.

Examples:

- `PaySimTransactionRow`
- future `OtherDatasetRow`

These belong in the ML training pipeline only.

### Layer 2: Canonical fraud feature schema

Create one stable feature model used by both:

- offline model training
- online model inference

Examples of canonical features:

- transaction amount
- transaction type category
- event hour / day index
- sender / receiver entity type
- transaction velocity counters
- first-time receiver flag
- cross-border or geography-derived flags
- device novelty
- derived temporal and behavioral statistics

This is the center of the design.

### Layer 3: Online inference schema

Keep the current product/event contracts stable.

Current path:

- UI input
- `transaction-ingestor`
- `feature-extractor`
- `transactions.enriched`
- `ml-service`

The ML service should map `transactions.enriched` into the canonical fraud feature schema instead of expecting raw dataset columns.

## What Changes In The Product Chain

We may add a small number of upstream fields where they materially improve feature quality, but we should avoid mirroring raw PaySim columns directly in the product API.

Potential additions worth evaluating:

- normalized transaction subtype aligned with PaySim categories
- explicit event hour / day bucket
- sender / destination entity-type flags
- additional behavioral counters that help recall

Fields that should stay out of user-facing product payloads unless clearly justified:

- raw dataset IDs like `nameOrig` / `nameDest`
- dataset labels like `isFraud`
- dataset-specific supervision flags like `isFlaggedFraud`
- leakage-prone balance fields copied straight from training CSV logic

## Training Plan

### Dataset source

Use PaySim as the primary public benchmark dataset.

The current repo-generated synthetic dataset remains useful only for:

- local smoke tests
- fallback demos
- very small dev loops

It should not remain the main benchmark dataset.

### Split strategy

Do not rely on random split only.

Use time-aware splits based on PaySim `step`.

Recommended starting split:

- train: earliest 70%
- validation: next 15%
- test: latest 15%

This is better than shuffled rows because SentinelPay is a fraud pipeline that predicts future events, not random reordered history.

### Modeling order

1. logistic regression baseline
2. gradient boosting baseline
3. deep learning only if it materially improves results

Possible advanced models later:

- MLP for tabular features
- temporal sequence model if event-history windows are introduced

Do not adopt deep learning for prestige alone.

## Evaluation Plan

### Primary offline evaluation

Use the held-out PaySim test split.

Primary metrics:

- recall
- PR-AUC
- precision at operating thresholds
- confusion matrix
- false positive rate
- calibration

Accuracy is not a headline metric due to severe class imbalance.

### Secondary robustness evaluation

After PaySim baseline works:

- map another payment / fraud dataset into the canonical feature schema
- run zero- or low-change evaluation to see whether the feature design overfits PaySim

This is optional at first, but good for credibility.

### Runtime evaluation

Keep end-to-end system tests separate from offline model evaluation.

Use generated or replayed transactions for:

- Kafka path validation
- enrichment path validation
- model scoring presence
- orchestration and alerting behavior
- UI sanity checks

Random runtime transactions are good for system smoke, not enough for model benchmarking.

## Implementation Phases

### Phase 1: Dataset ingestion

- add PaySim raw schema
- add loader for Kaggle-exported CSV
- add schema validation
- keep large raw files out of git

### Phase 2: Canonical feature layer

- define canonical fraud feature schema
- build PaySim -> canonical mapper
- build `transactions.enriched` -> canonical mapper
- add feature manifest output

### Phase 3: Training refactor

- refactor training scripts to consume canonical feature frames
- replace repo-generated dataset as main benchmark input
- preserve tiny local synthetic path for smoke/dev

### Phase 4: Evaluation artifacts

- store split metadata
- store metrics reports
- store model metadata including dataset source and feature schema version

### Phase 5: Runtime alignment

- evaluate whether `feature-extractor` needs a few more explicit derived fields
- keep UI/product payloads product-oriented, not dataset-oriented

## Acceptance Criteria

The change is in a good state when:

- PaySim can be loaded and validated locally
- the training pipeline can build canonical features from PaySim
- the inference path can build the same canonical features from `transactions.enriched`
- offline evaluation runs on train / validation / test splits
- model metrics are reported with recall and PR-AUC
- the system smoke tests still pass after ML integration changes

## Recommended GitHub Issue Scope

The implementation issue should be scoped as a dataset and feature-pipeline refactor, not “replace everything with Kaggle schema”.

Good scope:

- adopt PaySim as the benchmark dataset
- define canonical fraud feature schema
- refactor training and inference around that schema
- preserve current product contracts unless a small additive field change is justified
