import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const successRate = new Rate('success_rate');

const BASE_URL = __ENV.TARGET_URL || 'http://localhost:8081/api/v1/transactions';
const MODE = (__ENV.MODE || 'ramp').toLowerCase();

const currencyRates = {
  USD: 1,
  VND: 24000,
  EUR: 0.92,
  JPY: 150,
};

export const options = {
  scenarios: MODE === 'burst'
    ? {
        burst_mode: {
          executor: 'constant-arrival-rate',
          rate: 3000,
          timeUnit: '1s',
          duration: '20s',
          preAllocatedVUs: 400,
          maxVUs: 4000,
        },
      }
    : {
        ramp_mode: {
          executor: 'ramping-arrival-rate',
          timeUnit: '1s',
          preAllocatedVUs: 400,
          maxVUs: 4000,
          stages: [
            { target: 200, duration: '30s' },
            { target: 200, duration: '30s' },
            { target: 800, duration: '60s' },
            { target: 800, duration: '60s' },
            { target: 2000, duration: '60s' },
            { target: 2000, duration: '60s' },
          ],
        },
      },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<150'],
    success_rate: ['rate>0.99'],
  },
};

function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function pickCurrency() {
  const currencies = ['USD', 'VND', 'EUR', 'JPY'];
  return currencies[randomInt(0, currencies.length - 1)];
}

function pickAmountUsd() {
  const r = Math.random();
  if (r < 0.3) return Math.random() * 19 + 1; // 1-20
  if (r < 0.7) return Math.random() * 280 + 20; // 20-300
  return Math.random() * 1700 + 300; // 300-2000
}

function amountInCurrency(currency) {
  const usd = pickAmountUsd();
  const rate = currencyRates[currency] || 1;
  return Math.round(usd * rate);
}

function randomDeviceId(devicePool) {
  if (devicePool.length > 0 && Math.random() < 0.6) {
    return devicePool[randomInt(0, devicePool.length - 1)];
  }
  const id = `device-${uuidv4().slice(0, 8)}`;
  devicePool.push(id);
  return id;
}

export default function () {
  if (!globalThis.__devicePool) {
    globalThis.__devicePool = [];
  }

  const currency = pickCurrency();
  const payload = {
    transactionId: uuidv4(),
    type: 'P2P_TRANSFER',
    senderUserId: randomInt(1000, 5000),
    receiverUserId: randomInt(2000, 9000),
    merchantId: null,
    amount: amountInCurrency(currency),
    currency,
    deviceId: randomDeviceId(globalThis.__devicePool),
    timestamp: new Date().toISOString(),
  };

  const res = http.post(BASE_URL, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'fraud_ingest' },
  });

  const ok = check(res, {
    'status is 2xx/3xx': (r) => r.status >= 200 && r.status < 400,
  });
  successRate.add(ok);

  sleep(0.001);
}

export function handleSummary(data) {
  const totalReqs = data.metrics.http_reqs?.values?.count || 0;
  const avgLatency = data.metrics.http_req_duration?.values?.avg || 0;
  const p95Latency = data.metrics.http_req_duration?.values?.['p(95)'] || 0;
  const achievedRate = data.metrics.http_reqs?.values?.rate || 0;

  const summary = [
    '=== Fraud Ingest Load Test Summary ===',
    `Mode: ${MODE === 'burst' ? 'burst' : 'ramp'}`,
    `Total requests sent: ${totalReqs}`,
    `Achieved max throughput (req/sec): ${achievedRate.toFixed(2)}`,
    `Average latency (ms): ${avgLatency.toFixed(2)}`,
    `P95 latency (ms): ${p95Latency.toFixed(2)}`,
  ].join('\n');

  return {
    stdout: `${summary}\n`,
  };
}
