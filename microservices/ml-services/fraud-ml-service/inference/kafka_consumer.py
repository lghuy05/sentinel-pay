import json
import os

from confluent_kafka import Consumer


def get_env(name: str, default: str) -> str:
    value = os.getenv(name)
    return value if value is not None and value != "" else default


def create_consumer() -> tuple[Consumer, str]:
    bootstrap = get_env("KAFKA_BOOTSTRAP_SERVERS", "localhost:19092")
    input_topic = get_env("INPUT_TOPIC", "fraud.rules")
    group_id = get_env("KAFKA_GROUP_ID", "fraud-ml-service")

    consumer = Consumer(
        {
            "bootstrap.servers": bootstrap,
            "group.id": group_id,
            "auto.offset.reset": "earliest",
            "enable.auto.commit": True,
            "fetch.min.bytes": 1,
            "fetch.wait.max.ms": 25,
        }
    )
    consumer.subscribe([input_topic])
    return consumer, input_topic


def decode_message(raw_message) -> tuple[str | None, dict | None, str | None]:
    if raw_message is None:
        return None, None, None
    if raw_message.error():
        return None, None, str(raw_message.error())

    key = raw_message.key().decode("utf-8") if raw_message.key() else None
    value_bytes = raw_message.value()
    if not value_bytes:
        return key, None, None
    try:
        value = json.loads(value_bytes.decode("utf-8"))
    except json.JSONDecodeError:
        return key, None, "invalid-json"
    return key, value, None
