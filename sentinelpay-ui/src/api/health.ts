import axios from "axios";

export type HealthState = "up" | "down" | "unknown";

export interface ServiceHealth {
  name: string;
  url?: string;
  status: HealthState;
  lastChecked?: string;
  details?: string;
}

const nowIso = () => new Date().toISOString();

const services: ServiceHealth[] = [
  {
    name: "transaction-ingestor",
    url: import.meta.env.VITE_TRANSACTION_INGESTOR_URL || "/health/transaction-ingestor"
  },
  {
    name: "feature-extractor",
    url: import.meta.env.VITE_FEATURE_EXTRACTOR_URL || "/health/feature-extractor"
  },
  {
    name: "rule-engine",
    url: import.meta.env.VITE_RULE_ENGINE_URL || "/health/rule-engine"
  },
  {
    name: "blacklist-service",
    url: import.meta.env.VITE_BLACKLIST_SERVICE_URL || "/health/blacklist-service"
  },
  {
    name: "fraud-orchestrator",
    url: import.meta.env.VITE_FRAUD_ORCHESTRATOR_URL || "/health/fraud-orchestrator"
  },
  {
    name: "alert-service",
    url: import.meta.env.VITE_ALERT_SERVICE_URL || "/health/alert-service"
  },
  {
    name: "account-service",
    url: import.meta.env.VITE_ACCOUNT_SERVICE_URL || "/health/account-service"
  },
  {
    name: "ml-service",
    url: import.meta.env.VITE_ML_SERVICE_URL || "/health/ml-service"
  }
];

const fetchHealth = async (service: ServiceHealth): Promise<ServiceHealth> => {
  if (!service.url) {
    return {
      ...service,
      status: "unknown",
      lastChecked: nowIso(),
      details: "No health endpoint configured"
    };
  }

  try {
    const { data } = await axios.get(service.url, {
      timeout: 3000
    });

    return {
      ...service,
      status: data?.status === "UP" ? "up" : "down",
      lastChecked: nowIso(),
      details: data?.status || "UNKNOWN"
    };
  } catch (error) {
    const message = error instanceof Error ? error.message : "Health check failed";
    return {
      ...service,
      status: "down",
      lastChecked: nowIso(),
      details: message
    };
  }
};

export const fetchAllHealth = async () => {
  return Promise.all(services.map(fetchHealth));
};

export const serviceCatalog = services.map((service) => ({
  name: service.name,
  url: service.url
}));
