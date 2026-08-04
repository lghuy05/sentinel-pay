import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const gatewayTarget = "http://localhost:18082";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api/v1/transactions": {
        target: gatewayTarget,
        changeOrigin: true
      },
      "/api/decisions": {
        target: gatewayTarget,
        changeOrigin: true
      },
      "/api/feedback": {
        target: gatewayTarget,
        changeOrigin: true
      },
      "/api/v1/accounts": {
        target: gatewayTarget,
        changeOrigin: true
      },
      "/system": {
        target: gatewayTarget,
        changeOrigin: true
      },
      "/ml": {
        target: gatewayTarget,
        changeOrigin: true
      },
      "/health/transaction-ingestor": {
        target: gatewayTarget,
        changeOrigin: true
      },
      "/health/feature-extractor": {
        target: gatewayTarget,
        changeOrigin: true
      },
      "/health/rule-engine": {
        target: gatewayTarget,
        changeOrigin: true
      },
      "/health/blacklist-service": {
        target: gatewayTarget,
        changeOrigin: true
      },
      "/health/fraud-orchestrator": {
        target: gatewayTarget,
        changeOrigin: true
      },
      "/health/alert-service": {
        target: gatewayTarget,
        changeOrigin: true
      },
      "/health/account-service": {
        target: gatewayTarget,
        changeOrigin: true
      },
      "/health/ml-service": {
        target: gatewayTarget,
        changeOrigin: true
      }
    }
  }
});
