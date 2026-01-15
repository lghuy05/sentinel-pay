import json
import logging
import os
from datetime import datetime, timezone

from kafka import KafkaConsumer, KafkaProducer

LOG = logging.getLogger("fraud-ml-service")


def get_env(name, default):
    value = os.getenv(name)
    return value if value is not None and value != "" else default


def parse_bool(value, fallback=False):
    if value is None:
        return fallback
    if isinstance(value, bool):
        return value
    return str(value).lower() in ("1", "true", "yes", "y")


def compute_score(event):
    score = 0.1

    tx_count_1m = float(event.get("txCountLast1Min", 0) or 0)
    if tx_count_1m >= 5:
        score += 0.25
    elif tx_count_1m >= 3:
        score += 0.15

    tx_amount_1h = float(event.get("txAmountLast1Hour", 0) or 0)
    if tx_amount_1h >= 5_000_000:
        score += 0.25
    elif tx_amount_1h >= 1_000_000:
        score += 0.15

    currency = (event.get("currency") or "VND").upper()
    country = (event.get("country") or "VN").upper()
    if currency != "VND" or country != "VN":
        score += 0.2

    is_new_device = parse_bool(event.get("isNewDevice"), parse_bool(event.get("newDevice")))
    if is_new_device:
        score += 0.2

    if score > 1.0:
        score = 1.0
    if score < 0.0:
        score = 0.0
    return round(score, 4)


def build_output(event, score, model_version):
    return {
        "transactionId": event.get("transactionId"),
        "mlScore": score,
        "modelVersion": model_version,
        "evaluatedAt": datetime.now(timezone.utc).isoformat()
    }


def main():
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")

    bootstrap = get_env("KAFKA_BOOTSTRAP_SERVERS", "localhost:19092")
    input_topic = get_env("INPUT_TOPIC", "transactions.enriched")
    output_topic = get_env("OUTPUT_TOPIC", "fraud.ml")
    group_id = get_env("KAFKA_GROUP_ID", "fraud-ml-service")
    model_version = get_env("MODEL_VERSION", "baseline-v0.1")

    consumer = KafkaConsumer(
        input_topic,
        bootstrap_servers=bootstrap,
        group_id=group_id,
        auto_offset_reset="earliest",
        enable_auto_commit=True,
        value_deserializer=lambda v: json.loads(v.decode("utf-8")),
        key_deserializer=lambda v: v.decode("utf-8") if v else None,
    )

    producer = KafkaProducer(
        bootstrap_servers=bootstrap,
        value_serializer=lambda v: json.dumps(v).encode("utf-8"),
        key_serializer=lambda v: v.encode("utf-8") if v else None,
    )

    LOG.info("ML service started input=%s output=%s", input_topic, output_topic)

    for message in consumer:
        event = message.value
        if not isinstance(event, dict):
            LOG.warning("Skipping non-JSON payload: %s", message.value)
            continue

        score = compute_score(event)
        output = build_output(event, score, model_version)
        producer.send(output_topic, key=output["transactionId"], value=output)
        LOG.info("Scored txId=%s mlScore=%.4f", output["transactionId"], output["mlScore"])


if __name__ == "__main__":
    main()
