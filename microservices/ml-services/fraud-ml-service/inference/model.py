from __future__ import annotations

import math
from dataclasses import dataclass
from datetime import datetime
from typing import Any

import joblib


@dataclass
class FraudModel:
    model: Any
    feature_order: list[str]
    model_version: str
    trained_at: str | None = None
    metrics: dict[str, float] | None = None
    dataset_size: int | None = None

    @classmethod
    def load(cls, path: str) -> "FraudModel":
        artifact = joblib.load(path)
        return cls(
            model=artifact["model"],
            feature_order=list(artifact["feature_order"]),
            model_version=artifact.get("model_version", "logreg-synthetic-v1"),
            trained_at=artifact.get("trained_at"),
            metrics=artifact.get("metrics", {}),
            dataset_size=artifact.get("dataset_size"),
        )

    def score(self, event: dict[str, Any]) -> float:
        features = self._extract_features(event)
        proba = self.model.predict_proba([features])[0][1]
        return float(proba)

    def _extract_features(self, event: dict[str, Any]) -> list[float]:
        getters = {
            "amount": self._get_amount,
            "amount_normalized": self._get_amount_normalized,
            "log_amount_normalized": self._get_log_amount_normalized,
            "hour_of_day": self._get_hour_of_day,
            "day_index": self._get_day_index,
            "is_transfer": self._get_is_transfer,
            "is_cash_out": self._get_is_cash_out,
            "is_cash_in": self._get_is_cash_in,
            "is_payment": self._get_is_payment,
            "is_debit": self._get_is_debit,
            "origin_is_merchant": self._get_origin_is_merchant,
            "destination_is_merchant": self._get_destination_is_merchant,
            "origin_balance_before_normalized": self._get_origin_balance_before_normalized,
            "destination_balance_before_normalized": self._get_destination_balance_before_normalized,
            "amount_over_origin_balance": self._get_amount_over_origin_balance,
            "origin_balance_missing_or_zero": self._get_origin_balance_missing_or_zero,
            "destination_balance_missing_or_zero": self._get_destination_balance_missing_or_zero,
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
    def _get_amount_normalized(event: dict[str, Any]) -> float:
        value = event.get("amount_usd_equivalent", event.get("amountUsdEquivalent", event.get("amount", 0.0)))
        return FraudModel._to_float(value)

    @staticmethod
    def _get_log_amount_normalized(event: dict[str, Any]) -> float:
        amount = FraudModel._get_amount_normalized(event)
        return math.log1p(amount) if amount > 0 else 0.0

    @staticmethod
    def _parse_event_time(event: dict[str, Any]) -> datetime | None:
        timestamp = event.get("eventTime") or event.get("receivedAt")
        if not timestamp:
            return None
        try:
            return datetime.fromisoformat(str(timestamp).replace("Z", "+00:00"))
        except ValueError:
            return None

    @staticmethod
    def _get_hour_of_day(event: dict[str, Any]) -> int:
        parsed = FraudModel._parse_event_time(event)
        return parsed.hour if parsed else 0

    @staticmethod
    def _get_day_index(event: dict[str, Any]) -> int:
        parsed = FraudModel._parse_event_time(event)
        if not parsed:
            return 0
        return parsed.toordinal()

    @staticmethod
    def _normalized_type(event: dict[str, Any]) -> str:
        raw = event.get("type")
        return str(raw).upper() if raw is not None else ""

    @staticmethod
    def _get_is_transfer(event: dict[str, Any]) -> int:
        return 1 if FraudModel._normalized_type(event) in {"TRANSFER", "P2P_TRANSFER"} else 0

    @staticmethod
    def _get_is_cash_out(event: dict[str, Any]) -> int:
        return 1 if FraudModel._normalized_type(event) == "CASH_OUT" else 0

    @staticmethod
    def _get_is_cash_in(event: dict[str, Any]) -> int:
        return 1 if FraudModel._normalized_type(event) == "CASH_IN" else 0

    @staticmethod
    def _get_is_payment(event: dict[str, Any]) -> int:
        return 1 if FraudModel._normalized_type(event) in {"PAYMENT", "MERCHANT_PAYMENT"} else 0

    @staticmethod
    def _get_is_debit(event: dict[str, Any]) -> int:
        return 1 if FraudModel._normalized_type(event) == "DEBIT" else 0

    @staticmethod
    def _get_origin_is_merchant(event: dict[str, Any]) -> int:
        return 0

    @staticmethod
    def _get_destination_is_merchant(event: dict[str, Any]) -> int:
        merchant_id = event.get("merchantId")
        tx_type = FraudModel._normalized_type(event)
        return 1 if merchant_id not in (None, "", 0) or tx_type == "MERCHANT_PAYMENT" else 0

    @staticmethod
    def _get_origin_balance_before_normalized(event: dict[str, Any]) -> float:
        return FraudModel._to_float(event.get("senderBalanceMinor", event.get("sender_balance_minor", 0.0)))

    @staticmethod
    def _get_destination_balance_before_normalized(event: dict[str, Any]) -> float:
        return FraudModel._to_float(event.get("receiverBalanceMinor", event.get("receiver_balance_minor", 0.0)))

    @staticmethod
    def _get_amount_over_origin_balance(event: dict[str, Any]) -> float:
        amount = FraudModel._get_amount_normalized(event)
        origin_balance = FraudModel._get_origin_balance_before_normalized(event)
        if origin_balance <= 0:
            return 0.0
        return amount / origin_balance

    @staticmethod
    def _get_origin_balance_missing_or_zero(event: dict[str, Any]) -> int:
        return 1 if FraudModel._get_origin_balance_before_normalized(event) <= 0 else 0

    @staticmethod
    def _get_destination_balance_missing_or_zero(event: dict[str, Any]) -> int:
        return 1 if FraudModel._get_destination_balance_before_normalized(event) <= 0 else 0

    @staticmethod
    def _get_tx_count_1min(event: dict[str, Any]) -> Any:
        return event.get("tx_count_1min", event.get("tx_count_1m", event.get("txCountLast1Min")))

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
        parsed = FraudModel._parse_event_time(event)
        if not parsed:
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
