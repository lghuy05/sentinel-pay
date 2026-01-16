import argparse
import csv
import random
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import List, Tuple

VN_PEAK_HOURS = list(range(7, 22))
US_PEAK_HOURS = list(range(8, 23))
FRAUD_HOURS = [0, 1, 2, 3, 4]

CROSS_BORDER_PAIRS = [
    ("VN", "US"),
    ("US", "VN"),
    ("VN", "SG"),
    ("VN", "JP"),
    ("US", "CA"),
    ("US", "MX"),
]

VND_BUCKETS = [
    (0.65, (1_000, 200_000)),
    (0.20, (200_000, 2_000_000)),
    (0.10, (2_000_000, 10_000_000)),
    (0.04, (10_000_000, 50_000_000)),
    (0.01, (50_000_000, 120_000_000)),
]

USD_BUCKETS = [
    (0.60, (1, 50)),
    (0.25, (50, 300)),
    (0.10, (300, 1000)),
    (0.04, (1000, 5000)),
    (0.01, (5000, 12000)),
]


@dataclass
class TransactionRow:
    transaction_id: str
    sender_user_id: int
    sender_country: str
    receiver_country: str
    sender_currency: str
    amount: float
    timestamp: str
    expected_risk_label: str
    reason: str
    amount_risk_tier: str
    amount_usd_equivalent: float


def weighted_choice(rng: random.Random, buckets: List[Tuple[float, Tuple[int, int]]]) -> Tuple[int, int]:
    roll = rng.random()
    cumulative = 0.0
    for weight, bounds in buckets:
        cumulative += weight
        if roll <= cumulative:
            return bounds
    return buckets[-1][1]


def sample_amount(rng: random.Random, currency: str, cross_border: bool) -> float:
    if currency == "VND":
        low, high = weighted_choice(rng, VND_BUCKETS)
        if cross_border:
            high = min(high, 20_000_000)
        return float(rng.uniform(low, high))
    low, high = weighted_choice(rng, USD_BUCKETS)
    if cross_border:
        high = min(high, 2500)
    return float(rng.uniform(low, high))


def amount_risk_tier(currency: str, amount: float) -> str:
    if currency == "USD":
        if amount <= 50:
            return "LOW"
        if amount <= 300:
            return "MEDIUM"
        if amount <= 1000:
            return "HIGH"
        return "CRITICAL"
    if amount <= 200_000:
        return "LOW"
    if amount <= 2_000_000:
        return "MEDIUM"
    if amount <= 10_000_000:
        return "HIGH"
    return "CRITICAL"


def sample_sender_country(rng: random.Random, cross_border: bool) -> str:
    if cross_border:
        return rng.choice(["VN", "US"])
    return "VN" if rng.random() < 0.55 else "US"


def pick_receiver_country(rng: random.Random, sender_country: str, cross_border: bool) -> str:
    if not cross_border:
        return sender_country
    pair = rng.choice(CROSS_BORDER_PAIRS)
    if pair[0] == sender_country:
        return pair[1]
    return "US" if sender_country == "VN" else "VN"


def sample_timestamp(rng: random.Random, sender_country: str, fraud_cluster: bool) -> datetime:
    now = datetime.now(timezone.utc)
    day_offset = rng.randint(0, 29)
    base_date = now - timedelta(days=day_offset)

    if fraud_cluster:
        hour = rng.choice(FRAUD_HOURS)
    else:
        peak_hours = VN_PEAK_HOURS if sender_country == "VN" else US_PEAK_HOURS
        if rng.random() < 0.8:
            hour = rng.choice(peak_hours)
        else:
            hour = rng.randint(0, 23)

    minute = rng.randint(0, 59)
    second = rng.randint(0, 59)

    if sender_country == "VN":
        local = datetime(base_date.year, base_date.month, base_date.day, hour, minute, second, tzinfo=timezone(timedelta(hours=7)))
    elif sender_country == "US":
        local = datetime(base_date.year, base_date.month, base_date.day, hour, minute, second, tzinfo=timezone(timedelta(hours=-5)))
    else:
        local = datetime(base_date.year, base_date.month, base_date.day, hour, minute, second, tzinfo=timezone.utc)

    return local.astimezone(timezone.utc)


def risk_label(currency: str, amount: float, cross_border: bool, late_night: bool, burst: bool) -> Tuple[str, str]:
    tier = amount_risk_tier(currency, amount)

    if cross_border:
        if currency == "USD" and amount > 3000:
            return "HIGH", "cross-border high USD amount"
        if currency == "VND" and amount > 70_000_000:
            return "HIGH", "cross-border high VND amount"

    if tier == "CRITICAL" and (cross_border or late_night):
        return "HIGH", "critical amount during risky window"

    if burst or late_night:
        return "MEDIUM", "late-night or burst activity"

    if tier == "HIGH" and cross_border:
        return "MEDIUM", "elevated cross-border amount"

    return "LOW", "typical wallet behavior"


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate realistic wallet transaction dataset.")
    parser.add_argument("--rows", type=int, default=80_000)
    parser.add_argument("--seed", type=int, default=42)
    parser.add_argument(
        "--output",
        type=Path,
        default=Path(__file__).resolve().parent / "wallet_dataset.csv",
    )
    args = parser.parse_args()

    rng = random.Random(args.seed)

    normal_users_vn = [1000 + i for i in range(3000)]
    normal_users_us = [6000 + i for i in range(2500)]
    suspicious_users = [90000 + i for i in range(200)]

    contacts = {}

    rows: List[TransactionRow] = []

    for idx in range(args.rows):
        cross_border = rng.random() < 0.05
        sender_country = sample_sender_country(rng, cross_border)
        receiver_country = pick_receiver_country(rng, sender_country, cross_border)

        currency = "VND" if sender_country == "VN" else "USD"

        sender_user_id = rng.choice(normal_users_vn if sender_country == "VN" else normal_users_us)
        burst = False
        if rng.random() < 0.06:
            sender_user_id = rng.choice(suspicious_users)
            burst = True

        receiver_user_id = rng.randint(20000, 50000)
        contact_key = (sender_user_id, receiver_user_id)
        first_contact = contact_key not in contacts
        contacts[contact_key] = True

        amount = sample_amount(rng, currency, cross_border)
        tier = amount_risk_tier(currency, amount)

        fraud_cluster = burst or (cross_border and rng.random() < 0.15)
        timestamp = sample_timestamp(rng, sender_country, fraud_cluster)
        local_hour = timestamp.astimezone(timezone(timedelta(hours=7 if sender_country == "VN" else -5))).hour
        late_night = local_hour in FRAUD_HOURS

        label, reason = risk_label(currency, amount, cross_border, late_night, burst)
        if first_contact and label == "LOW":
            label = "MEDIUM"
            reason = "first-time recipient"

        amount_usd = amount if currency == "USD" else amount / 25_000
        rows.append(
            TransactionRow(
                transaction_id=f"tx-{idx:06d}",
                sender_user_id=sender_user_id,
                sender_country=sender_country,
                receiver_country=receiver_country,
                sender_currency=currency,
                amount=round(amount, 2),
                timestamp=timestamp.isoformat(),
                expected_risk_label=label,
                reason=reason,
                amount_risk_tier=tier,
                amount_usd_equivalent=round(amount_usd, 2),
            )
        )

    args.output.parent.mkdir(parents=True, exist_ok=True)
    with args.output.open("w", newline="") as csvfile:
        writer = csv.writer(csvfile)
        writer.writerow([
            "transactionId",
            "senderUserId",
            "senderCountry",
            "receiverCountry",
            "senderCurrency",
            "amount",
            "amountUsdEquivalent",
            "timestamp",
            "expectedRiskLabel",
            "reason",
            "amountRiskTier",
        ])
        for row in rows:
            writer.writerow([
                row.transaction_id,
                row.sender_user_id,
                row.sender_country,
                row.receiver_country,
                row.sender_currency,
                row.amount,
                row.amount_usd_equivalent,
                row.timestamp,
                row.expected_risk_label,
                row.reason,
                row.amount_risk_tier,
            ])

    print(f"Saved {len(rows)} rows to {args.output}")


if __name__ == "__main__":
    main()
