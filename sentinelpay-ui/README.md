# SentinelPay UI

Internal operations console for the SentinelPay distributed fraud detection system.

## Features

- Simulate transactions without curl
- Review recent ingested transactions
- Visualize fraud decisions stored by fraud-orchestrator
- Monitor service health via `/actuator/health`
- Inspect rule/blacklist matches and score breakdowns

## Requirements

- Node.js 18+
- SentinelPay backend running (see root `scripts/start-services.sh`)

## Run locally

```bash
cd sentinelpay-ui
npm install
npm run dev
```

Open `http://localhost:5173`.

## API configuration

By default, the UI calls the backend on `http://localhost` with these ports:

- transaction-ingestor: `8081`
- feature-extractor: `8082`
- rule-engine: `8083`
- blacklist-service: `8084`
- fraud-orchestrator: `8085`
- alert-service: `8086`

The dev server proxies requests to avoid CORS issues. Default proxies:

- `/api/v1/transactions` → `http://localhost:8081`
- `/api/decisions` → `http://localhost:8085`
- `/api/feedback` → `http://localhost:8085`
- `/health/*` → service `/actuator/health`

Override any endpoint with environment variables:

```bash
VITE_API_BASE_URL=http://localhost:8081
VITE_TRANSACTIONS_API_BASE_URL=http://localhost:8081
VITE_DECISIONS_API_BASE_URL=http://localhost:8085
VITE_TRANSACTION_INGESTOR_URL=/health/transaction-ingestor
VITE_FEATURE_EXTRACTOR_URL=/health/feature-extractor
VITE_RULE_ENGINE_URL=/health/rule-engine
VITE_BLACKLIST_SERVICE_URL=/health/blacklist-service
VITE_FRAUD_ORCHESTRATOR_URL=/health/fraud-orchestrator
VITE_ALERT_SERVICE_URL=/health/alert-service
VITE_ML_SERVICE_URL=
VITE_KAFKA_HEALTH_URL=
VITE_REDIS_HEALTH_URL=
VITE_POSTGRES_HEALTH_URL=
```

Set `VITE_API_BASE_URL` if you want a single base URL for `/api/v1/*` requests.

## Demo flow

1. Start SentinelPay services.
2. Open the UI and go to **Simulate Transaction**.
3. Create a transaction and submit it (transaction-ingestor accepts it immediately).
4. Review the **Transaction History** table to confirm it is stored.
5. Wait for the pipeline to emit a decision.
6. Go to **Fraud Decisions** and refresh.
7. Click a row to inspect rule/blacklist matches and score breakdowns.

## Notes

- `System Status` pings `/actuator/health`. Services without a health endpoint will show as `Unknown` until configured.
- The UI does not mock fraud decisions. All records come from `GET /decisions` on fraud-orchestrator via `/api/decisions` proxy.
