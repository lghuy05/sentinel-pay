#!/bin/bash

echo "🚀 Starting SentinelPay Infrastructure..."

# Create network if not exists
docker network create sentinelpay-network 2>/dev/null || true

echo "📊 Starting Databases..."
cd database
docker-compose up -d

echo "📨 Starting Kafka..."
cd ../kafka-setup
docker-compose up -d

echo "📈 Starting Monitoring..."
cd ../prometheus-grafana
docker-compose up -d

echo ""
echo "✅ INFRASTRUCTURE READY:"
echo "   PostgreSQL:     localhost:15432 (sentinel/sentinel123)"
echo "   Redis:          localhost:16379"
echo "   Kafka:          localhost:19092"
echo "   Kafka UI:       http://localhost:8080"
echo "   Prometheus:     http://localhost:9090"
echo "   Grafana:        http://localhost:3000 (admin/admin123)"
echo ""
echo "⏳ Waiting for Kafka to be ready..."
sleep 15

echo "📋 Creating Kafka topics..."
docker exec sentinelpay-kafka kafka-topics --create \
  --topic sentinelpay.transactions.raw \
  --bootstrap-server localhost:19092 \
  --partitions 3 \
  --replication-factor 1

docker exec sentinelpay-kafka kafka-topics --create \
  --topic sentinelpay.transactions.enriched \
  --bootstrap-server localhost:19092 \
  --partitions 3 \
  --replication-factor 1

docker exec sentinelpay-kafka kafka-topics --list \
  --bootstrap-server localhost:19092
