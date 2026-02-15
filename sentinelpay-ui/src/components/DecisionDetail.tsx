import { useMemo } from "react";
import type { DecisionRecord, FraudDecision } from "../api/decisions";

const DecisionDetail = ({ decision }: { decision: DecisionRecord | null }) => {
  const decisionClass = (decisionValue?: FraudDecision) => {
    if (decisionValue === "ALLOW") return "table-chip allow";
    if (decisionValue === "BLOCK") return "table-chip block";
    return "table-chip hold";
  };

  const decisionIcon = (decisionValue?: FraudDecision) => {
    if (decisionValue === "ALLOW") return "pi pi-check";
    if (decisionValue === "BLOCK") return "pi pi-ban";
    return "pi pi-clock";
  };

  const formatScore = (value?: number) => (value == null ? "-" : value.toFixed(3));
  const formatDate = (value?: string) => (value ? new Date(value).toLocaleString() : "-");

  const explanationText = useMemo(() => {
    if (!decision) return "";
    if (decision.finalDecision === "BLOCK") {
      return decision.decisionReason
        ? `Blocked due to ${decision.decisionReason}.`
        : "This transaction exceeded the risk threshold. Review rule and ML signals before releasing funds.";
    }
    if (decision.finalDecision === "HOLD") {
      return decision.decisionReason
        ? `Held for review due to ${decision.decisionReason}.`
        : "Transaction flagged for manual review based on combined risk signals.";
    }
    return decision.decisionReason
      ? `Allowed due to ${decision.decisionReason}.`
      : "Transaction cleared through the risk stack. Monitor downstream alerts if needed.";
  }, [decision]);

  const featureSnapshot = useMemo(() => {
    if (!decision?.featuresJson) return "{}";
    try {
      return JSON.stringify(JSON.parse(decision.featuresJson), null, 2);
    } catch {
      return decision.featuresJson || "{}";
    }
  }, [decision]);

  if (!decision) {
    return (
      <div className="surface-card">
        <h3 className="section-title">Decision Detail</h3>
        <p className="subtle">Select a decision to inspect the score breakdown and matched signals.</p>
      </div>
    );
  }

  return (
    <div className="surface-card">
      <h3 className="section-title">Decision Detail</h3>
      <div style={{ display: "flex", gap: 16, flexWrap: "wrap" }}>
        <div style={{ flex: 1, minWidth: 220 }}>
          <p className="subtle">Transaction</p>
          <h3>{decision.transactionId}</h3>
          <p className="subtle">Decision</p>
          <div className={decisionClass(decision.finalDecision)}>
            <i className={decisionIcon(decision.finalDecision)}></i>
            {decision.finalDecision}
          </div>
        </div>
        <div style={{ flex: 1, minWidth: 220 }}>
          <p className="subtle">Score Breakdown</p>
          <ul style={{ listStyle: "none", padding: 0, margin: 0 }}>
            <li>
              Rule score: <strong>{formatScore(decision.ruleScore)}</strong>
            </li>
            <li>
              Rule band: <strong>{decision.ruleBand || "-"}</strong>
            </li>
            <li>
              ML score: <strong>{formatScore(decision.mlScore)}</strong>
            </li>
            <li>
              ML band: <strong>{decision.mlBand || "-"}</strong>
            </li>
            <li>
              Blacklist hit: <strong>{decision.blacklistHit ? "YES" : "NO"}</strong>
            </li>
          </ul>
        </div>
        <div style={{ flex: 1, minWidth: 220 }}>
          <p className="subtle">Explanation</p>
          <p>{explanationText}</p>
          <p className="subtle">Decided At</p>
          <strong>{formatDate(decision.createdAt)}</strong>
        </div>
      </div>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
          gap: 16,
          marginTop: 16
        }}
      >
        <div>
          <p className="subtle">Rule Matches</p>
          {decision.ruleMatches?.length ? (
            <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
              {(decision.ruleMatches as string[]).map((rule) => (
                <span key={rule} className="badge warn">
                  {rule}
                </span>
              ))}
            </div>
          ) : (
            <p className="subtle">No rule matches</p>
          )}
        </div>
        <div>
          <p className="subtle">Decision Reason</p>
          <p>{decision.decisionReason || "-"}</p>
        </div>
        <div>
          <p className="subtle">Features</p>
          <pre style={{ maxHeight: 180, overflow: "auto" }}>{featureSnapshot}</pre>
        </div>
        <div>
          <p className="subtle">Model Info</p>
          <p>Version: {decision.modelVersion || "-"}</p>
          <p>Rule version: {decision.ruleVersion ?? "-"}</p>
        </div>
      </div>
    </div>
  );
};

export default DecisionDetail;
