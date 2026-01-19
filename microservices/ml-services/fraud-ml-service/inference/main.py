import json
import logging
import os
import sys
import threading
from datetime import datetime, timezone
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path

from kafka_consumer import create_consumer, decode_message
from kafka_producer import create_producer, serialize_message
from model import FraudModel


def get_env(name: str, default: str) -> str:
    value = os.getenv(name)
    return value if value is not None and value != "" else default


class MlState:
    def __init__(self, model_path: str) -> None:
        self.model_path = model_path
        self.model = FraudModel.load(model_path)
        self.training_status = "idle"
        self.training_error = None
        self.last_trained_at = self.model.trained_at
        self.metrics = self.model.metrics or {}
        self.dataset_size = self.model.dataset_size
        self.lock = threading.Lock()

    def reload(self) -> None:
        with self.lock:
            self.model = FraudModel.load(self.model_path)
            self.last_trained_at = self.model.trained_at
            self.metrics = self.model.metrics or {}
            self.dataset_size = self.model.dataset_size

    def status(self) -> dict:
        with self.lock:
            return {
                "status": "UP",
                "model_version": self.model.model_version,
                "trained_at": self.last_trained_at,
                "metrics": self.metrics,
                "dataset_size": self.dataset_size,
                "training_status": self.training_status,
                "training_error": self.training_error,
            }


class HealthHandler(BaseHTTPRequestHandler):
    def do_GET(self) -> None:
        if self.path == "/health/ml-service":
            self._send_json({"status": "UP"})
            return
        if self.path == "/ml/status":
            self._send_json(STATE.status())
            return
        self.send_response(404)
        self.end_headers()

    def do_POST(self) -> None:
        if self.path == "/ml/reload":
            try:
                STATE.reload()
                self._send_json({"status": "reloaded", "model_version": STATE.model.model_version})
            except Exception as exc:
                self._send_json({"status": "error", "error": str(exc)}, status=500)
            return
        if self.path == "/ml/retrain":
            if STATE.training_status == "running":
                self._send_json({"status": "running"})
                return
            thread = threading.Thread(target=run_training, daemon=True)
            thread.start()
            self._send_json({"status": "started"})
            return
        self.send_response(404)
        self.end_headers()

    def _send_json(self, payload: dict, status: int = 200) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format: str, *args) -> None:
        return


def run_training() -> None:
    STATE.training_status = "running"
    STATE.training_error = None
    try:
        base_dir = Path(__file__).resolve().parents[1]
        if str(base_dir) not in sys.path:
            sys.path.insert(0, str(base_dir))
        from training.train import train_and_save

        result = train_and_save()
        if result and result.get("model_path"):
            STATE.model_path = result["model_path"]
            STATE.reload()
        STATE.training_status = "completed"
    except Exception as exc:
        STATE.training_status = "failed"
        STATE.training_error = str(exc)


def start_health_server() -> None:
    port = int(get_env("HEALTH_PORT", "8091"))
    server = ThreadingHTTPServer(("0.0.0.0", port), HealthHandler)
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    logging.info("Health endpoint started on :%s", port)


def extract_features(event: dict) -> dict:
    if "features" in event and isinstance(event["features"], dict):
        return event["features"]
    return event


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    start_health_server()

    consumer, input_topic = create_consumer()
    producer, output_topic = create_producer()

    logging.info("Fraud ML worker started input=%s output=%s", input_topic, output_topic)

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

        rule_band = str(event.get("ruleBand") or event.get("rule_band") or "").upper()
        if rule_band != "GRAY":
            continue

        transaction_id = event.get("transactionId") or key
        if not transaction_id:
            logging.warning("Skipping event without transactionId: %s", event)
            continue

        features = extract_features(event)
        with STATE.lock:
            score = STATE.model.score(features)
            model_version = STATE.model.model_version

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
    model_path = get_env("MODEL_PATH", str(Path(__file__).resolve().parents[1] / "model_artifacts" / "fraud_model.joblib"))
    STATE = MlState(model_path)
    main()
