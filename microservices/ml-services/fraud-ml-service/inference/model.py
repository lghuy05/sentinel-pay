from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime
from typing import Any

import joblib


@dataclass
class FraudModel:
    model: Any
    feature_order: list[str]
    model_version: str

    @classmethod
    def load(cls, path: str) -> "FraudModel":
        artifact = joblib.load(path)
        return cls(
            model=artifact["model"],
            feature_order=list(artifact["feature_order"]),
            model_version=artifact.get("model_version", "logreg-synthetic-v1"),
        )

    def score(self, event: dict[str, Any]) -> float:
        features = self._extract_features(event)
        proba = self.model.predict_proba([features])[0][1]
        return float(proba)

    def _extract_features(self, event: dict[str, Any]) -> list[float]:
        getters = {
            "amount": self._get_amount,
            "tx_count_1min": self._get_tx_count_1min,
            "tx_amount_1hour": self._get_tx_amount_1hour,
            "is_new_device": self._get_is_new_device,
            "is_overseas": self._get_is_overseas,
            "is_night": self._get_is_night,
            "is_cross_border": self._get_is_cross_border,
            "amount_usd_equivalent": self._get_amount_usd_equivalent,
            "avg_amount_usd_24h": self._get_avg_amount_usd_24h,
            "amount_risk_tier": self._get_amount_risk_tier,
            "sender_account_age_days": self._get_sender_account_age_days,
            "receiver_account_age_days": self._get_receiver_account_age_days,
            "sender_tx_count_24h": self._get_sender_tx_count_24h,
            "sender_total_amount_usd_24h": self._get_sender_total_amount_usd_24h,
            "receiver_inbound_count_24h": self._get_receiver_inbound_count_24h,
            "sender_receiver_tx_count_24h": self._get_sender_receiver_tx_count_24h,
            "small_amount_burst_1m": self._get_small_amount_burst_1m,
            "small_amount_spread_24h": self._get_small_amount_spread_24h,
            "is_first_time_receiver": self._get_is_first_time_receiver,
        }

        features: list[float] = []
        for name in self.feature_order:
            getter = getters.get(name)
            value = getter(event) if getter else 0
            features.append(self._to_float(value))
        return features

    @staticmethod
    def _to_float(value: Any) -> float:
        if value is None:
            return 0.0
        if isinstance(value, bool):
            return 1.0 if value else 0.0
        try:
            return float(value)
        except (TypeError, ValueError):
            return 0.0

    @staticmethod
    def _get_amount(event: dict[str, Any]) -> Any:
        return event.get("amount")

    @staticmethod
    def _get_tx_count_1min(event: dict[str, Any]) -> Any:
        return event.get("tx_count_1min", event.get("txCountLast1Min"))

    @staticmethod
    def _get_tx_amount_1hour(event: dict[str, Any]) -> Any:
        return event.get("tx_amount_1hour", event.get("txAmountLast1Hour"))

    @staticmethod
    def _get_is_new_device(event: dict[str, Any]) -> Any:
        return event.get("is_new_device", event.get("isNewDevice", event.get("newDevice")))

    @staticmethod
    def _get_is_overseas(event: dict[str, Any]) -> int:
        if "is_overseas" in event:
            return event.get("is_overseas")
        if "overseas" in event:
            return event.get("overseas")
        return 0

    @staticmethod
    def _get_is_night(event: dict[str, Any]) -> int:
        if "is_night" in event:
            return event.get("is_night")
        timestamp = event.get("eventTime") or event.get("receivedAt")
        if not timestamp:
            return 0
        try:
            ts = str(timestamp).replace("Z", "+00:00")
            parsed = datetime.fromisoformat(ts)
        except ValueError:
            return 0
        hour = parsed.hour
        return 1 if hour >= 22 or hour <= 5 else 0

    @staticmethod
    def _get_is_cross_border(event: dict[str, Any]) -> Any:
        return event.get("is_cross_border", event.get("crossBorder"))

    @staticmethod
    def _get_amount_usd_equivalent(event: dict[str, Any]) -> Any:
        return event.get("amountUsdEquivalent", event.get("amount_usd_equivalent"))

    @staticmethod
    def _get_amount_risk_tier(event: dict[str, Any]) -> Any:
        raw = event.get("amount_risk_tier", event.get("amountRiskTier"))
        if raw is None:
            return 0
        if isinstance(raw, (int, float)):
            return raw
        tier = str(raw).upper()
        if tier == "LOW":
            return 1
        if tier == "MEDIUM":
            return 2
        if tier == "HIGH":
            return 3
        if tier == "CRITICAL":
            return 4
        return 0

    @staticmethod
    def _get_sender_account_age_days(event: dict[str, Any]) -> Any:
        return event.get("senderAccountAgeDays")

    @staticmethod
    def _get_receiver_account_age_days(event: dict[str, Any]) -> Any:
        return event.get("receiverAccountAgeDays")

    @staticmethod
    def _get_sender_tx_count_24h(event: dict[str, Any]) -> Any:
        return event.get("senderTxCount24h")

    @staticmethod
    def _get_sender_total_amount_usd_24h(event: dict[str, Any]) -> Any:
        return event.get("senderTotalAmountUsd24h")

    @staticmethod
    def _get_receiver_inbound_count_24h(event: dict[str, Any]) -> Any:
        return event.get("receiverInboundCount24h")

    @staticmethod
    def _get_sender_receiver_tx_count_24h(event: dict[str, Any]) -> Any:
        return event.get("senderReceiverTxCount24h", event.get("sender_receiver_tx_count_24h"))

    @staticmethod
    def _get_avg_amount_usd_24h(event: dict[str, Any]) -> Any:
        total = event.get("senderTotalAmountUsd24h")
        count = event.get("senderTxCount24h")
        try:
            total_value = float(total) if total is not None else 0.0
            count_value = float(count) if count is not None else 0.0
        except (TypeError, ValueError):
            return 0.0
        if count_value <= 0:
            return 0.0
        return total_value / count_value

    @staticmethod
    def _get_small_amount_burst_1m(event: dict[str, Any]) -> Any:
        tx_count = event.get("tx_count_1min", event.get("txCountLast1Min"))
        amount = event.get("amountUsdEquivalent", event.get("amount_usd_equivalent"))
        try:
            tx_count = float(tx_count) if tx_count is not None else 0.0
            amount = float(amount) if amount is not None else 0.0
        except (TypeError, ValueError):
            return 0
        return 1 if tx_count >= 5 and amount <= 15 else 0

    @staticmethod
    def _get_small_amount_spread_24h(event: dict[str, Any]) -> Any:
        avg = FraudModel._get_avg_amount_usd_24h(event)
        tx_count = event.get("senderTxCount24h")
        try:
            tx_count_value = float(tx_count) if tx_count is not None else 0.0
            avg_value = float(avg) if avg is not None else 0.0
        except (TypeError, ValueError):
            return 0
        return 1 if tx_count_value >= 30 and avg_value <= 20 else 0

    @staticmethod
    def _get_is_first_time_receiver(event: dict[str, Any]) -> Any:
        return event.get("is_first_time_receiver", event.get("firstTimeContact"))
