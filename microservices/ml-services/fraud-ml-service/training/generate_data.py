import argparse
from pathlib import Path

import numpy as np
import pandas as pd

FEATURES = [
    "amount",
    "tx_count_1min",
    "tx_amount_1hour",
    "is_new_device",
    "is_overseas",
    "is_night",
]

EXTRA_FIELDS = [
    "currency",
    "sender_country",
    "receiver_country",
    "is_cross_border",
    "amount_risk_tier",
    "amount_usd_equivalent",
    "avg_amount_usd_24h",
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


def build_dataset(rows: int, seed: int) -> pd.DataFrame:
    rng = np.random.default_rng(seed)

    is_overseas = rng.binomial(1, 0.14, rows)
    is_new_device = rng.binomial(1, 0.18, rows)
    is_night = rng.binomial(1, 0.22, rows)

    amount_base = rng.lognormal(mean=12.6, sigma=0.75, size=rows)
    amount = amount_base * (1 + is_overseas * rng.uniform(1.6, 4.8, size=rows))
    amount = np.clip(amount, 5_000, 30_000_000)

    tx_count_1min = rng.poisson(lam=1.3 + 2.4 * is_overseas + 1.3 * is_new_device)
    tx_count_1min = np.clip(tx_count_1min, 0, 18)

    tx_amount_1hour = amount * rng.uniform(0.9, 6.0, size=rows) + rng.normal(0, 120_000, size=rows)
    tx_amount_1hour = np.clip(tx_amount_1hour, 0, 80_000_000)

    overseas_origin = rng.binomial(1, 0.5, rows)
    sender_country = np.where(is_overseas == 1, np.where(overseas_origin == 1, "VN", "US"), "VN")
    receiver_country = np.where(
        is_overseas == 1,
        np.where(sender_country == "VN", "US", "VN"),
        "VN",
    )
    is_cross_border = (sender_country != receiver_country).astype(int)
    currency = np.where(sender_country == "US", "USD", "VND")
    amount_risk_tier = []
    for value, curr in zip(amount, currency):
        if curr == "USD":
            if value <= 50:
                amount_risk_tier.append("LOW")
            elif value <= 300:
                amount_risk_tier.append("MEDIUM")
            elif value <= 1000:
                amount_risk_tier.append("HIGH")
            else:
                amount_risk_tier.append("CRITICAL")
        else:
            if value <= 200_000:
                amount_risk_tier.append("LOW")
            elif value <= 2_000_000:
                amount_risk_tier.append("MEDIUM")
            elif value <= 10_000_000:
                amount_risk_tier.append("HIGH")
            else:
                amount_risk_tier.append("CRITICAL")

    amount_usd_equivalent = np.where(currency == "USD", amount, amount / 25_000)

    sender_account_age_days = rng.integers(1, 1500, rows)
    receiver_account_age_days = rng.integers(1, 800, rows)
    sender_tx_count_24h = rng.poisson(lam=2.5 + 1.0 * is_overseas + 0.5 * is_new_device, size=rows)
    sender_tx_count_24h = np.clip(sender_tx_count_24h, 0, 120)
    sender_total_amount_usd_24h = amount_usd_equivalent * rng.uniform(0.5, 3.5, size=rows)
    receiver_inbound_count_24h = rng.poisson(lam=1.5 + 0.8 * is_overseas, size=rows)
    receiver_inbound_count_24h = np.clip(receiver_inbound_count_24h, 0, 80)
    avg_amount_usd_24h = sender_total_amount_usd_24h / np.maximum(sender_tx_count_24h, 1)
    small_amount_burst_1m = ((tx_count_1min >= 5) & (amount_usd_equivalent <= 15)).astype(int)
    small_amount_spread_24h = ((sender_tx_count_24h >= 30) & (avg_amount_usd_24h <= 20)).astype(int)
    sender_receiver_tx_count_24h = rng.poisson(lam=1.5 + 2.5 * small_amount_spread_24h, size=rows)
    sender_receiver_tx_count_24h = np.minimum(sender_tx_count_24h, sender_receiver_tx_count_24h)
    sender_receiver_tx_count_24h = np.clip(sender_receiver_tx_count_24h, 0, 15)
    is_first_time_receiver = rng.binomial(1, 0.2 + 0.15 * is_new_device, rows)

    high_velocity = (tx_count_1min >= 5).astype(int)
    amount_million = amount / 1_000_000
    hour_amount_factor = tx_amount_1hour / 5_000_000

    # Risk score with interaction effects for overseas, night, new device, and bursty micro-transfers.
    risk = (
        -3.4
        + 0.55 * amount_million
        + 0.35 * hour_amount_factor
        + 0.9 * is_overseas
        + 0.7 * is_new_device
        + 0.5 * is_night
        + 0.55 * (sender_tx_count_24h / 40)
        + 0.45 * (receiver_inbound_count_24h / 25)
        + 0.7 * (sender_receiver_tx_count_24h / 5)
        + 1.3 * small_amount_burst_1m
        + 1.1 * small_amount_spread_24h
        + 1.0 * (is_overseas * is_new_device)
        + 0.9 * (is_overseas * is_night)
        + 0.8 * (is_new_device * is_night)
        + 0.7 * high_velocity
    )

    fraud_prob = 1 / (1 + np.exp(-risk))
    fraud = rng.binomial(1, np.clip(fraud_prob, 0.0, 0.98))

    data = {
        "amount": amount.round(2),
        "tx_count_1min": tx_count_1min.astype(int),
        "tx_amount_1hour": tx_amount_1hour.round(2),
        "is_new_device": is_new_device.astype(int),
        "is_overseas": is_overseas.astype(int),
        "is_night": is_night.astype(int),
        "currency": currency,
        "sender_country": sender_country,
        "receiver_country": receiver_country,
        "is_cross_border": is_cross_border.astype(int),
        "amount_risk_tier": amount_risk_tier,
        "amount_usd_equivalent": amount_usd_equivalent.round(2),
        "avg_amount_usd_24h": avg_amount_usd_24h.round(2),
        "sender_account_age_days": sender_account_age_days.astype(int),
        "receiver_account_age_days": receiver_account_age_days.astype(int),
        "sender_tx_count_24h": sender_tx_count_24h.astype(int),
        "sender_total_amount_usd_24h": sender_total_amount_usd_24h.round(2),
        "receiver_inbound_count_24h": receiver_inbound_count_24h.astype(int),
        "sender_receiver_tx_count_24h": sender_receiver_tx_count_24h.astype(int),
        "small_amount_burst_1m": small_amount_burst_1m.astype(int),
        "small_amount_spread_24h": small_amount_spread_24h.astype(int),
        "is_first_time_receiver": is_first_time_receiver.astype(int),
        "fraud": fraud.astype(int),
    }

    return pd.DataFrame(data, columns=FEATURES + EXTRA_FIELDS + ["fraud"])


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate synthetic fraud dataset.")
    parser.add_argument("--rows", type=int, default=10_000)
    parser.add_argument("--seed", type=int, default=42)
    parser.add_argument(
        "--output",
        type=Path,
        default=Path(__file__).resolve().parent / "dataset.csv",
    )
    args = parser.parse_args()

    df = build_dataset(args.rows, args.seed)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    df.to_csv(args.output, index=False)
    print(f"Saved {len(df)} rows to {args.output}")


if __name__ == "__main__":
    main()
