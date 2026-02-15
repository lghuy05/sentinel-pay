import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api/v1/transactions": {
        target: "http://localhost:8081",
        changeOrigin: true
      },
      "/api/decisions": {
        target: "http://localhost:8085",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, "")
      },
      "/api/feedback": {
        target: "http://localhost:8085",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, "")
      },
      "/api/v1/accounts": {
        target: "http://localhost:8087",
        changeOrigin: true
      },
      "/health/transaction-ingestor": {
        target: "http://localhost:8081",
        changeOrigin: true
      },
      "/health/feature-extractor": {
        target: "http://localhost:8082",
        changeOrigin: true
      },
      "/health/rule-engine": {
        target: "http://localhost:8083",
        changeOrigin: true
      },
      "/health/blacklist-service": {
        target: "http://localhost:8084",
        changeOrigin: true
      },
      "/health/fraud-orchestrator": {
        target: "http://localhost:8085",
        changeOrigin: true
      },
      "/health/alert-service": {
        target: "http://localhost:8086",
        changeOrigin: true
      },
      "/health/account-service": {
        target: "http://localhost:8087",
        changeOrigin: true
      },
      "/health/ml-service": {
        target: "http://localhost:18091",
        changeOrigin: true
      }
    }
  }
});
