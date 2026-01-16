import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_DECISIONS_API_BASE_URL || import.meta.env.VITE_API_BASE_URL || ""
});

export type FraudDecision = "ALLOW" | "BLOCK" | "HOLD";

export interface DecisionRecord {
  id: number;
  transactionId: string;
  decision: FraudDecision;
  finalScore: number;
  mlScore: number;
  ruleScore: number;
  blacklistScore: number;
  ruleMatches: string[] | string;
  blacklistMatches: string[] | string;
  riskScore?: number;
  riskLevel?: string;
  triggeredRules?: string[] | string;
  hardStopMatches?: string[] | string;
  hardStopDecision?: string | null;
  decidedAt: string;
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
  ruleMatches: normalizeList(record.ruleMatches),
  blacklistMatches: normalizeList(record.blacklistMatches),
  triggeredRules: normalizeList(record.triggeredRules),
  hardStopMatches: normalizeList(record.hardStopMatches)
});

export const fetchDecisions = async (limit = 50) => {
  const { data } = await api.get<DecisionRecord[]>("/api/v1/decisions", {
    params: { limit }
  });
  return data.map(normalizeDecision);
};

export const fetchDecision = async (transactionId: string) => {
  const { data } = await api.get<DecisionRecord>(`/api/v1/decisions/${transactionId}`);
  return normalizeDecision(data);
};
