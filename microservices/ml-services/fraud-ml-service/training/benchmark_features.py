from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Any

import numpy as np
import pandas as pd

FEATURE_ORDER = [
    "amount_normalized",
    "log_amount_normalized",
    "hour_of_day",
    "day_index",
    "is_transfer",
    "is_cash_out",
    "is_cash_in",
    "is_payment",
    "is_debit",
    "origin_is_merchant",
    "destination_is_merchant",
    "origin_balance_before_normalized",
    "destination_balance_before_normalized",
    "amount_over_origin_balance",
    "origin_balance_missing_or_zero",
    "destination_balance_missing_or_zero",
]


@dataclass
class DatasetBundle:
    dataset_type: str
    frame: pd.DataFrame
    labels: pd.Series
    row_ids: pd.Series
    sort_key: pd.Series | None = None


def load_supported_dataset(path: Path, max_rows: int | None = None) -> DatasetBundle:
    df = pd.read_csv(path, nrows=max_rows)
    columns = set(df.columns)

    if {"step", "type", "amount", "nameOrig", "nameDest", "isFraud"}.issubset(columns):
        return build_paysim_bundle(df)
    if {"amount", "amount_usd_equivalent", "fraud"}.issubset(columns):
        return build_synthetic_bundle(df)

    raise ValueError(
        "Unsupported dataset schema. Expected PaySim columns or the repo synthetic training schema."
    )


def build_paysim_bundle(df: pd.DataFrame) -> DatasetBundle:
    work = df.copy()

    work["type"] = work["type"].astype(str).str.upper()
    work["step"] = pd.to_numeric(work["step"], errors="coerce").fillna(0).astype(int)
    work["amount"] = pd.to_numeric(work["amount"], errors="coerce").fillna(0.0)
    work["oldbalanceOrg"] = pd.to_numeric(work["oldbalanceOrg"], errors="coerce").fillna(0.0)
    work["oldbalanceDest"] = pd.to_numeric(work["oldbalanceDest"], errors="coerce").fillna(0.0)
    work["isFraud"] = pd.to_numeric(work["isFraud"], errors="coerce").fillna(0).astype(int)

    features = pd.DataFrame(index=work.index)
    features["amount_normalized"] = work["amount"]
    features["log_amount_normalized"] = np.log1p(work["amount"])
    features["hour_of_day"] = work["step"] % 24
    features["day_index"] = work["step"] // 24
    features["is_transfer"] = (work["type"] == "TRANSFER").astype(int)
    features["is_cash_out"] = (work["type"] == "CASH_OUT").astype(int)
    features["is_cash_in"] = (work["type"] == "CASH_IN").astype(int)
    features["is_payment"] = (work["type"] == "PAYMENT").astype(int)
    features["is_debit"] = (work["type"] == "DEBIT").astype(int)
    features["origin_is_merchant"] = work["nameOrig"].astype(str).str.startswith("M").astype(int)
    features["destination_is_merchant"] = work["nameDest"].astype(str).str.startswith("M").astype(int)
    features["origin_balance_before_normalized"] = work["oldbalanceOrg"]
    features["destination_balance_before_normalized"] = work["oldbalanceDest"]
    features["amount_over_origin_balance"] = np.where(
        work["oldbalanceOrg"] > 0,
        work["amount"] / work["oldbalanceOrg"],
        0.0,
    )
    features["origin_balance_missing_or_zero"] = (work["oldbalanceOrg"] <= 0).astype(int)
    features["destination_balance_missing_or_zero"] = (work["oldbalanceDest"] <= 0).astype(int)
    features = features.fillna(0.0)

    row_ids = work.get("nameOrig", pd.Series([f"row-{idx}" for idx in work.index], index=work.index)).astype(str) + "->" + work.get("nameDest", pd.Series(["unknown"] * len(work), index=work.index)).astype(str) + "#" + work.index.astype(str)
    return DatasetBundle(
        dataset_type="paysim",
        frame=features[FEATURE_ORDER],
        labels=work["isFraud"].astype(int),
        row_ids=row_ids,
        sort_key=work["step"],
    )


def build_synthetic_bundle(df: pd.DataFrame) -> DatasetBundle:
    work = df.copy()
    for column in ["amount_usd_equivalent", "fraud"]:
        work[column] = pd.to_numeric(work[column], errors="coerce").fillna(0.0)

    features = pd.DataFrame(index=work.index)
    amount_normalized = pd.to_numeric(work.get("amount_usd_equivalent", 0.0), errors="coerce").fillna(0.0)
    features["amount_normalized"] = amount_normalized
    features["log_amount_normalized"] = np.log1p(amount_normalized)
    features["hour_of_day"] = 0
    features["day_index"] = 0
    features["is_transfer"] = 0
    features["is_cash_out"] = 0
    features["is_cash_in"] = 0
    features["is_payment"] = 0
    features["is_debit"] = 0
    features["origin_is_merchant"] = 0
    features["destination_is_merchant"] = 0
    features["origin_balance_before_normalized"] = 0.0
    features["destination_balance_before_normalized"] = 0.0
    features["amount_over_origin_balance"] = 0.0
    features["origin_balance_missing_or_zero"] = 1
    features["destination_balance_missing_or_zero"] = 1
    features = features.fillna(0.0)

    row_ids = pd.Series([f"synthetic-{idx}" for idx in work.index], index=work.index)
    return DatasetBundle(
        dataset_type="synthetic",
        frame=features[FEATURE_ORDER],
        labels=work["fraud"].astype(int),
        row_ids=row_ids,
        sort_key=None,
    )
