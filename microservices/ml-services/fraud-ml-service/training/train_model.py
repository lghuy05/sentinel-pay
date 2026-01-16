import argparse
from datetime import datetime, timezone
from pathlib import Path

import joblib
import pandas as pd
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import accuracy_score, roc_auc_score
from sklearn.model_selection import train_test_split

FEATURE_ORDER = [
    "amount",
    "tx_count_1min",
    "tx_amount_1hour",
    "is_new_device",
    "is_overseas",
    "is_night",
    "is_cross_border",
    "amount_usd_equivalent",
    "avg_amount_usd_24h",
    "amount_risk_tier",
    "sender_account_age_days",
    "receiver_account_age_days",
    "sender_tx_count_24h",
    "sender_total_amount_usd_24h",
    "receiver_inbound_count_24h",
    "sender_receiver_tx_count_24h",
    "small_amount_burst_1m",
    "small_amount_spread_24h",
    "is_first_time_receiver",
]

MODEL_VERSION = "logreg-synthetic-v3"


def main() -> None:
    parser = argparse.ArgumentParser(description="Train fraud model.")
    parser.add_argument(
        "--data",
        type=Path,
        default=Path(__file__).resolve().parent / "dataset.csv",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=Path(__file__).resolve().parents[1]
        / "model_artifacts"
        / "fraud_model.joblib",
    )
    args = parser.parse_args()

    df = pd.read_csv(args.data)
    X = df[FEATURE_ORDER].copy()
    if "amount_risk_tier" in X.columns:
        tier_map = {"LOW": 1, "MEDIUM": 2, "HIGH": 3, "CRITICAL": 4}
        X["amount_risk_tier"] = X["amount_risk_tier"].map(tier_map).fillna(0).astype(float)
    y = df["fraud"].astype(int)

    X_train, X_test, y_train, y_test = train_test_split(
        X,
        y,
        test_size=0.2,
        random_state=42,
        stratify=y,
    )

    model = LogisticRegression(max_iter=1000, solver="liblinear", C=5.0, class_weight="balanced")
    model.fit(X_train, y_train)

    preds = model.predict(X_test)
    proba = model.predict_proba(X_test)[:, 1]

    accuracy = accuracy_score(y_test, preds)
    roc_auc = roc_auc_score(y_test, proba)

    print(f"Accuracy: {accuracy:.4f}")
    print(f"ROC AUC: {roc_auc:.4f}")

    artifact = {
        "model": model,
        "feature_order": FEATURE_ORDER,
        "model_version": MODEL_VERSION,
        "trained_at": datetime.now(timezone.utc).isoformat(),
        "metrics": {"accuracy": accuracy, "roc_auc": roc_auc},
    }

    args.output.parent.mkdir(parents=True, exist_ok=True)
    joblib.dump(artifact, args.output)
    print(f"Saved model to {args.output}")


if __name__ == "__main__":
    main()
