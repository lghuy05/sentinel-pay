import json
import os

from confluent_kafka import Producer


def get_env(name: str, default: str) -> str:
    value = os.getenv(name)
    return value if value is not None and value != "" else default


def create_producer() -> tuple[Producer, str]:
    bootstrap = get_env("KAFKA_BOOTSTRAP_SERVERS", "localhost:19092")
    output_topic = get_env("OUTPUT_TOPIC", "fraud.ml")

    producer = Producer({"bootstrap.servers": bootstrap})
    return producer, output_topic


def serialize_message(key: str | None, value: dict) -> tuple[bytes | None, bytes]:
    key_bytes = key.encode("utf-8") if key else None
    value_bytes = json.dumps(value).encode("utf-8")
    return key_bytes, value_bytes
