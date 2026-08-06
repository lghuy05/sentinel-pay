from __future__ import annotations

import argparse
import json
from datetime import datetime, timezone
from pathlib import Path

import joblib
import pandas as pd
from sklearn.metrics import accuracy_score, average_precision_score, confusion_matrix, f1_score, precision_score, recall_score, roc_auc_score
from sklearn.model_selection import train_test_split

from benchmark_features import FEATURE_ORDER, DatasetBundle, load_supported_dataset
from model_candidates import build_model, supported_models


def split_bundle(bundle: DatasetBundle, train_ratio: float, val_ratio: float) -> tuple[pd.Index, pd.Index, pd.Index, str]:
    if bundle.dataset_type == "paysim" and bundle.sort_key is not None:
        order = bundle.sort_key.sort_values(kind="stable").index
        total = len(order)
        train_end = int(total * train_ratio)
        val_end = int(total * (train_ratio + val_ratio))
        return order[:train_end], order[train_end:val_end], order[val_end:], "time"

    train_idx, test_idx = train_test_split(
        bundle.frame.index,
        test_size=(1.0 - train_ratio),
        random_state=42,
        stratify=bundle.labels,
    )
    relative_val = val_ratio / (1.0 - train_ratio)
    val_idx, final_test_idx = train_test_split(
        test_idx,
        test_size=(1.0 - relative_val),
        random_state=42,
        stratify=bundle.labels.loc[test_idx],
    )
    return pd.Index(train_idx), pd.Index(val_idx), pd.Index(final_test_idx), "random-stratified"


def evaluate_split(model: object, X: pd.DataFrame, y: pd.Series, threshold: float) -> tuple[dict[str, float], pd.Series, pd.Series]:
    scores = pd.Series(model.predict_proba(X)[:, 1], index=X.index)
    preds = (scores >= threshold).astype(int)
    metrics = {
        "accuracy": float(accuracy_score(y, preds)),
        "precision": float(precision_score(y, preds, zero_division=0)),
        "recall": float(recall_score(y, preds, zero_division=0)),
        "f1": float(f1_score(y, preds, zero_division=0)),
        "roc_auc": float(roc_auc_score(y, scores)) if y.nunique() > 1 else 0.0,
        "pr_auc": float(average_precision_score(y, scores)) if y.nunique() > 1 else 0.0,
    }
    return metrics, scores, preds


def choose_threshold(y: pd.Series, scores: pd.Series) -> tuple[float, dict[str, float]]:
    best_threshold = 0.50
    best_metrics: dict[str, float] | None = None
    for threshold in [x / 100 for x in range(5, 100, 5)]:
        preds = (scores >= threshold).astype(int)
        metrics = {
            "accuracy": float(accuracy_score(y, preds)),
            "precision": float(precision_score(y, preds, zero_division=0)),
            "recall": float(recall_score(y, preds, zero_division=0)),
            "f1": float(f1_score(y, preds, zero_division=0)),
        }
        if best_metrics is None:
            best_threshold = threshold
            best_metrics = metrics
            continue
        candidate = (metrics["f1"], metrics["recall"], metrics["accuracy"])
        current = (best_metrics["f1"], best_metrics["recall"], best_metrics["accuracy"])
        if candidate > current:
            best_threshold = threshold
            best_metrics = metrics
    assert best_metrics is not None
    return best_threshold, best_metrics


def run_benchmark(
    data_path: Path,
    output_dir: Path,
    train_ratio: float,
    val_ratio: float,
    threshold: float | None,
    model_name: str,
    max_rows: int | None = None,
) -> dict[str, object]:
    bundle = load_supported_dataset(data_path, max_rows=max_rows)
    train_idx, val_idx, test_idx, split_strategy = split_bundle(bundle, train_ratio, val_ratio)

    X_train = bundle.frame.loc[train_idx]
    y_train = bundle.labels.loc[train_idx]
    X_val = bundle.frame.loc[val_idx]
    y_val = bundle.labels.loc[val_idx]
    X_test = bundle.frame.loc[test_idx]
    y_test = bundle.labels.loc[test_idx]

    model = build_model(model_name)
    model.fit(X_train, y_train)

    val_scores = pd.Series(model.predict_proba(X_val)[:, 1], index=X_val.index)
    selected_threshold = threshold
    threshold_source = "manual"
    if selected_threshold is None:
        selected_threshold, threshold_fit_metrics = choose_threshold(y_val, val_scores)
        threshold_source = "validation-f1"
    else:
        threshold_fit_metrics = {}

    val_metrics, _, _ = evaluate_split(model, X_val, y_val, selected_threshold)
    test_metrics, test_scores, test_preds = evaluate_split(model, X_test, y_test, selected_threshold)
    tn, fp, fn, tp = confusion_matrix(y_test, test_preds, labels=[0, 1]).ravel()

    output_dir.mkdir(parents=True, exist_ok=True)

    dataset_name = data_path.stem
    model_version = f"{model_name}-{bundle.dataset_type}-benchmark-v1"
    trained_at = datetime.now(timezone.utc).isoformat()

    artifact_path = output_dir / f"{dataset_name}-fraud-model.joblib"
    predictions_path = output_dir / f"{dataset_name}-predictions.csv"
    metrics_path = output_dir / f"{dataset_name}-metrics.json"

    artifact = {
        "model": model,
        "feature_order": FEATURE_ORDER,
        "model_version": model_version,
        "trained_at": trained_at,
        "dataset_type": bundle.dataset_type,
        "dataset_source": str(data_path),
        "split_strategy": split_strategy,
        "threshold": selected_threshold,
        "threshold_source": threshold_source,
        "model_name": model_name,
        "metrics": {
            "threshold_fit": threshold_fit_metrics,
            "validation": val_metrics,
            "test": test_metrics,
            "confusion_matrix": {"tn": int(tn), "fp": int(fp), "fn": int(fn), "tp": int(tp)},
        },
        "dataset_size": int(len(bundle.labels)),
        "max_rows": max_rows,
    }
    joblib.dump(artifact, artifact_path)

    prediction_frame = pd.DataFrame(
        {
            "rowId": bundle.row_ids.loc[test_idx].values,
            "actualLabel": y_test.values,
            "predictedScore": test_scores.values,
            "predictedLabel": test_preds.values,
            "detectedCorrectly": (test_preds.values == y_test.values).astype(int),
        }
    )
    prediction_frame.to_csv(predictions_path, index=False)

    metrics_payload = {
        "dataset_type": bundle.dataset_type,
        "dataset_source": str(data_path),
        "model_name": model_name,
        "split_strategy": split_strategy,
        "threshold": selected_threshold,
        "threshold_source": threshold_source,
        "feature_order": FEATURE_ORDER,
        "dataset_size": int(len(bundle.labels)),
        "max_rows": max_rows,
        "train_size": int(len(train_idx)),
        "validation_size": int(len(val_idx)),
        "test_size": int(len(test_idx)),
        "threshold_fit_metrics": threshold_fit_metrics,
        "validation_metrics": val_metrics,
        "test_metrics": test_metrics,
        "confusion_matrix": {"tn": int(tn), "fp": int(fp), "fn": int(fn), "tp": int(tp)},
        "correct_detection_percentage": round(float((test_preds.values == y_test.values).mean() * 100.0), 4),
    }
    metrics_path.write_text(json.dumps(metrics_payload, indent=2), encoding="utf-8")

    return {
        "dataset_type": bundle.dataset_type,
        "split_strategy": split_strategy,
        "threshold": selected_threshold,
        "threshold_source": threshold_source,
        "validation_metrics": val_metrics,
        "test_metrics": test_metrics,
        "correct_detection_percentage": metrics_payload["correct_detection_percentage"],
        "artifact_path": str(artifact_path),
        "predictions_path": str(predictions_path),
        "metrics_path": str(metrics_path),
    }


def main() -> None:
    parser = argparse.ArgumentParser(description="Train and evaluate SentinelPay ML benchmark dataset.")
    parser.add_argument("--data", type=Path, required=True, help="Path to PaySim CSV or supported benchmark CSV")
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path(__file__).resolve().parents[1] / "model_artifacts" / "benchmark",
    )
    parser.add_argument("--train-ratio", type=float, default=0.70)
    parser.add_argument("--val-ratio", type=float, default=0.15)
    parser.add_argument("--threshold", type=float, default=None)
    parser.add_argument("--model", choices=supported_models(), default="logreg")
    parser.add_argument("--max-rows", type=int, default=None)
    args = parser.parse_args()

    result = run_benchmark(
        data_path=args.data,
        output_dir=args.output_dir,
        train_ratio=args.train_ratio,
        val_ratio=args.val_ratio,
        threshold=args.threshold,
        model_name=args.model,
        max_rows=args.max_rows,
    )

    print(f"Dataset type: {result['dataset_type']}")
    print(f"Split strategy: {result['split_strategy']}")
    print(f"Selected threshold: {float(result['threshold']):.2f} ({result['threshold_source']})")
    print(f"Validation metrics: {json.dumps(result['validation_metrics'])}")
    print(f"Test metrics: {json.dumps(result['test_metrics'])}")
    print(f"Correct detection percentage: {float(result['correct_detection_percentage']):.2f}%")
    print(f"Saved artifact: {result['artifact_path']}")
    print(f"Saved predictions: {result['predictions_path']}")
    print(f"Saved metrics: {result['metrics_path']}")


if __name__ == "__main__":
    main()
