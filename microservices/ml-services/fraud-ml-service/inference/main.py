import json
import logging
import os
import threading
from datetime import datetime, timezone
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

from kafka_consumer import create_consumer, decode_message
from kafka_producer import create_producer, serialize_message
from model import FraudModel


def get_env(name: str, default: str) -> str:
    value = os.getenv(name)
    return value if value is not None and value != "" else default


class HealthHandler(BaseHTTPRequestHandler):
    def do_GET(self) -> None:
        if self.path != "/health/ml-service":
            self.send_response(404)
            self.end_headers()
            return

        payload = {"status": "UP"}
        body = json.dumps(payload).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format: str, *args) -> None:
        return


def start_health_server() -> None:
    port = int(get_env("HEALTH_PORT", "8091"))
    server = ThreadingHTTPServer(("0.0.0.0", port), HealthHandler)
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    logging.info("Health endpoint started on :%s/health/ml-service", port)


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    start_health_server()

    model_path = get_env("MODEL_PATH", "../model_artifacts/fraud_model.joblib")
    model = FraudModel.load(model_path)

    consumer, input_topic = create_consumer()
    producer, output_topic = create_producer()
    model_version = get_env("MODEL_VERSION", model.model_version)

    logging.info(
        "Fraud ML worker started model=%s input=%s output=%s",
        model_path,
        input_topic,
        output_topic,
    )

    while True:
        message = consumer.poll(1.0)
        if message is None:
            continue

        key, event, error = decode_message(message)
        if error:
            logging.warning("Kafka message error: %s", error)
            continue
        if not isinstance(event, dict):
            logging.warning("Skipping non-JSON payload: %s", event)
            continue

        transaction_id = event.get("transactionId") or key
        if not transaction_id:
            logging.warning("Skipping event without transactionId: %s", event)
            continue

        score = model.score(event)
        output = {
            "transactionId": transaction_id,
            "mlScore": round(score, 6),
            "modelVersion": model_version,
            "evaluatedAt": datetime.now(timezone.utc).isoformat(),
        }
        key_bytes, value_bytes = serialize_message(transaction_id, output)
        producer.produce(output_topic, key=key_bytes, value=value_bytes)
        producer.poll(0)
        logging.info("Scored txId=%s mlScore=%.6f", transaction_id, output["mlScore"])


if __name__ == "__main__":
    main()
