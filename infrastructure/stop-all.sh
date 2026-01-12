#!/bin/bash

echo "🛑 Stopping SentinelPay Infrastructure..."

cd prometheus-grafana
docker-compose down

cd ../kafka-setup
docker-compose down

cd ../database
docker-compose down

echo "✅ All infrastructure stopped."
