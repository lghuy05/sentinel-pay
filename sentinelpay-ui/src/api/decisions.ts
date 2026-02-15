import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_DECISIONS_API_BASE_URL || import.meta.env.VITE_API_BASE_URL || ""
});

export type FraudDecision = "ALLOW" | "BLOCK" | "HOLD";

export interface DecisionRecord {
  id: number;
  transactionId: string;
  accountId?: number;
  amount?: number;
  country?: string;
  featuresJson?: string;
  blacklistHit?: boolean;
  ruleScore?: number;
  ruleBand?: string;
  ruleMatches?: string[] | string;
  mlScore?: number;
  mlBand?: string;
  finalDecision?: FraudDecision;
  decisionReason?: string;
  modelVersion?: string;
  ruleVersion?: number;
  trueLabel?: boolean | null;
  reviewed?: boolean;
  createdAt?: string;
}

const normalizeList = (value: unknown): string[] => {
  if (!value) return [];
  if (Array.isArray(value)) {
    return value.map((item) => String(item));
  }
  if (typeof value === "string") {
    try {
      const parsed = JSON.parse(value);
      if (Array.isArray(parsed)) {
        return parsed.map((item) => String(item));
      }
    } catch {
      return value.split(",").map((item) => item.trim()).filter(Boolean);
    }
  }
  return [];
};

const normalizeDecision = (record: DecisionRecord): DecisionRecord => ({
  ...record,
  ruleMatches: normalizeList(record.ruleMatches)
});

export const fetchDecisions = async (limit = 50) => {
  const { data } = await api.get<DecisionRecord[]>("/api/decisions", {
    params: { limit }
  });
  return data.map(normalizeDecision);
};

export const fetchUnreviewedDecisions = async (limit = 100) => {
  const { data } = await api.get<DecisionRecord[]>("/api/decisions", {
    params: { limit, reviewed: false }
  });
  return data.map(normalizeDecision);
};

export const fetchDecision = async (transactionId: string) => {
  const { data } = await api.get<DecisionRecord>(`/api/decisions/${transactionId}`);
  return normalizeDecision(data);
};
