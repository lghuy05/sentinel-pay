from __future__ import annotations

from typing import Any

from sklearn.ensemble import HistGradientBoostingClassifier, RandomForestClassifier
from sklearn.linear_model import LogisticRegression
from xgboost import XGBClassifier


def build_model(model_name: str) -> Any:
    if model_name == "logreg":
        return LogisticRegression(
            max_iter=2000,
            solver="liblinear",
            class_weight="balanced",
            C=2.0,
        )
    if model_name == "random_forest":
        return RandomForestClassifier(
            n_estimators=300,
            max_depth=14,
            min_samples_leaf=3,
            class_weight="balanced_subsample",
            random_state=42,
            n_jobs=-1,
        )
    if model_name == "hist_gb":
        return HistGradientBoostingClassifier(
            max_iter=300,
            learning_rate=0.05,
            max_depth=8,
            min_samples_leaf=40,
            random_state=42,
        )
    if model_name == "xgboost":
        return XGBClassifier(
            n_estimators=500,
            max_depth=6,
            learning_rate=0.05,
            subsample=0.9,
            colsample_bytree=0.9,
            objective="binary:logistic",
            eval_metric="logloss",
            random_state=42,
            n_jobs=4,
        )
    raise ValueError(f"Unsupported model '{model_name}'")


def supported_models() -> list[str]:
    return ["logreg", "random_forest", "hist_gb", "xgboost"]
