from __future__ import annotations

import argparse
import json
from pathlib import Path

import pandas as pd

from evaluate_benchmark import run_benchmark
from model_candidates import supported_models


def main() -> None:
    parser = argparse.ArgumentParser(description="Compare benchmark models on a supported fraud dataset.")
    parser.add_argument("--data", type=Path, required=True, help="Path to PaySim CSV or supported benchmark CSV")
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path(__file__).resolve().parents[1] / "model_artifacts" / "benchmark-comparison",
    )
    parser.add_argument("--train-ratio", type=float, default=0.70)
    parser.add_argument("--val-ratio", type=float, default=0.15)
    parser.add_argument("--max-rows", type=int, default=None)
    args = parser.parse_args()

    rows: list[dict[str, object]] = []
    for model_name in supported_models():
        result = run_benchmark(
            data_path=args.data,
            output_dir=args.output_dir / model_name,
            train_ratio=args.train_ratio,
            val_ratio=args.val_ratio,
            threshold=None,
            model_name=model_name,
            max_rows=args.max_rows,
        )
        rows.append(
            {
                "model": model_name,
                "dataset_type": result["dataset_type"],
                "threshold": result["threshold"],
                "threshold_source": result["threshold_source"],
                "accuracy": result["test_metrics"]["accuracy"],
                "precision": result["test_metrics"]["precision"],
                "recall": result["test_metrics"]["recall"],
                "f1": result["test_metrics"]["f1"],
                "roc_auc": result["test_metrics"]["roc_auc"],
                "pr_auc": result["test_metrics"]["pr_auc"],
                "correct_detection_percentage": result["correct_detection_percentage"],
            }
        )

    summary = pd.DataFrame(rows).sort_values(
        by=["pr_auc", "recall", "f1", "accuracy"],
        ascending=[False, False, False, False],
        kind="stable",
    )
    args.output_dir.mkdir(parents=True, exist_ok=True)
    summary_csv = args.output_dir / f"{args.data.stem}-model-comparison.csv"
    summary_json = args.output_dir / f"{args.data.stem}-model-comparison.json"
    summary.to_csv(summary_csv, index=False)
    summary_json.write_text(json.dumps(rows, indent=2), encoding="utf-8")

    print(summary.to_string(index=False))
    print(f"Saved comparison CSV: {summary_csv}")
    print(f"Saved comparison JSON: {summary_json}")


if __name__ == "__main__":
    main()
