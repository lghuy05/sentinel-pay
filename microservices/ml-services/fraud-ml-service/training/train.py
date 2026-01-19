import json
import os
from datetime import datetime, timezone
from pathlib import Path

import joblib
import pandas as pd
import psycopg2
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import precision_score, recall_score, roc_auc_score
from sklearn.model_selection import train_test_split

FEATURE_ORDER = [
    "amount",
    "tx_count_1min",
    "tx_amount_1hour",
    "is_new_device",
    "is_overseas",
    "amount_usd_equivalent",
    "amount_risk_tier",
    "sender_account_age_days",
    "receiver_account_age_days",
    "sender_tx_count_24h",
    "sender_total_amount_usd_24h",
    "receiver_inbound_count_24h",
    "sender_receiver_tx_count_24h",
    "is_first_time_receiver",
]


def get_env(name: str, default: str) -> str:
    value = os.getenv(name)
    return value if value is not None and value != "" else default


def db_connection():
    return psycopg2.connect(
        host=get_env("DB_HOST", "localhost"),
        port=int(get_env("DB_PORT", "15432")),
        dbname=get_env("DB_NAME", "orchestrator_db"),
        user=get_env("DB_USER", "sentinel"),
        password=get_env("DB_PASSWORD", "sentinel123"),
    )


def load_training_data() -> tuple[pd.DataFrame, pd.Series]:
    query = """
        SELECT features_json, true_label
        FROM fraud_decisions
        WHERE true_label IS NOT NULL
          AND reviewed = true
          AND decision_reason LIKE 'ML_%'
          AND created_at >= NOW() - INTERVAL '30 days'
    """
    with db_connection() as conn:
        df = pd.read_sql(query, conn)
    if df.empty:
        raise RuntimeError("No labeled data available for training")

    features = df["features_json"].apply(lambda raw: json.loads(raw) if raw else {})
    feature_rows = []
    for row in features:
        feature_rows.append(row if isinstance(row, dict) else {})
    X = pd.DataFrame(feature_rows)
    rename_map = {
        "txCountLast1Min": "tx_count_1min",
        "tx_count_1m": "tx_count_1min",
        "txAmountLast1Hour": "tx_amount_1hour",
        "senderAccountAgeDays": "sender_account_age_days",
        "receiverAccountAgeDays": "receiver_account_age_days",
        "senderTxCount24h": "sender_tx_count_24h",
        "senderTotalAmountUsd24h": "sender_total_amount_usd_24h",
        "receiverInboundCount24h": "receiver_inbound_count_24h",
        "senderReceiverTxCount24h": "sender_receiver_tx_count_24h",
        "amountUsdEquivalent": "amount_usd_equivalent",
        "amountRiskTier": "amount_risk_tier",
        "firstTimeContact": "is_first_time_receiver",
        "newDevice": "is_new_device",
        "overseas": "is_overseas",
    }
    X = X.rename(columns=rename_map)
    for col in FEATURE_ORDER:
        if col not in X.columns:
            X[col] = 0.0
    X = X[FEATURE_ORDER].copy()
    X = X.fillna(0.0)
    y = df["true_label"].astype(int)
    return X, y


def next_model_version(conn) -> str:
    conn.cursor().execute(
        """
        CREATE TABLE IF NOT EXISTS model_registry (
            id SERIAL PRIMARY KEY,
            model_version TEXT NOT NULL,
            trained_at TIMESTAMP NOT NULL,
            metrics JSONB,
            dataset_size INTEGER NOT NULL
        )
        """
    )
    conn.commit()
    cursor = conn.cursor()
    cursor.execute("SELECT model_version FROM model_registry ORDER BY id DESC LIMIT 1")
    row = cursor.fetchone()
    if not row or not row[0].startswith("model_v"):
        return "model_v1"
    try:
        current = int(row[0].split("_v")[1])
    except Exception:
        return "model_v1"
    return f"model_v{current + 1}"


def train_and_save() -> dict:
    X, y = load_training_data()

    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.2, random_state=42, stratify=y
    )

    model = LogisticRegression(max_iter=1000, solver="liblinear", class_weight="balanced")
    model.fit(X_train, y_train)

    proba = model.predict_proba(X_test)[:, 1]
    preds = (proba >= 0.5).astype(int)
    metrics = {
        "precision": float(precision_score(y_test, preds)),
        "recall": float(recall_score(y_test, preds)),
        "roc_auc": float(roc_auc_score(y_test, proba)),
    }

    output_dir = Path(__file__).resolve().parents[1] / "model_artifacts"
    output_dir.mkdir(parents=True, exist_ok=True)
    model_path = output_dir / "fraud_model.joblib"

    with db_connection() as conn:
        version = next_model_version(conn)
        trained_at = datetime.now(timezone.utc).isoformat()
        artifact = {
            "model": model,
            "feature_order": FEATURE_ORDER,
            "model_version": version,
            "trained_at": trained_at,
            "metrics": metrics,
            "dataset_size": int(len(y)),
        }
        joblib.dump(artifact, model_path)

        cursor = conn.cursor()
        cursor.execute(
            "INSERT INTO model_registry (model_version, trained_at, metrics, dataset_size) VALUES (%s, %s, %s, %s)",
            (version, datetime.now(timezone.utc), json.dumps(metrics), len(y)),
        )
        conn.commit()

    return {"model_path": str(model_path), "model_version": version, "metrics": metrics}


if __name__ == "__main__":
    result = train_and_save()
    print(result)
