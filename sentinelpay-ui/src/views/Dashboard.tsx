import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Button } from "primereact/button";
import { DataTable } from "primereact/datatable";
import { Column } from "primereact/column";
import { InputSwitch } from "primereact/inputswitch";

import { fetchDecisions, type DecisionRecord, type FraudDecision } from "../api/decisions";

const Dashboard = () => {
  const [decisions, setDecisions] = useState<DecisionRecord[]>([]);
  const [lastRefresh, setLastRefresh] = useState("-");
  const [errorMessage, setErrorMessage] = useState("");
  const [autoRefresh, setAutoRefresh] = useState(true);
  const intervalRef = useRef<number | undefined>(undefined);

  const restoreScroll = (scrollX: number, scrollY: number) => {
    requestAnimationFrame(() => {
      window.scrollTo({ top: scrollY, left: scrollX, behavior: "auto" });
      requestAnimationFrame(() => {
        window.scrollTo({ top: scrollY, left: scrollX, behavior: "auto" });
      });
    });
  };

  const loadDecisions = useCallback(async () => {
    const scrollX = window.scrollX;
    const scrollY = window.scrollY;
    try {
      const data = await fetchDecisions();
      setDecisions(
        [...data].sort(
          (a, b) => new Date(b.createdAt || 0).getTime() - new Date(a.createdAt || 0).getTime()
        )
      );
      setErrorMessage("");
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to load decisions";
      setErrorMessage(message);
    } finally {
      setLastRefresh(new Date().toLocaleTimeString());
      restoreScroll(scrollX, scrollY);
    }
  }, []);

  const todayStart = useMemo(() => {
    const start = new Date();
    start.setHours(0, 0, 0, 0);
    return start;
  }, []);

  const decisionsToday = useMemo(
    () => decisions.filter((decision) => decision.createdAt && new Date(decision.createdAt) >= todayStart),
    [decisions, todayStart]
  );

  const totalToday = decisionsToday.length;
  const blockedToday = decisionsToday.filter((decision) => decision.finalDecision === "BLOCK").length;
  const allowedToday = decisionsToday.filter((decision) => decision.finalDecision === "ALLOW").length;
  const blockRate = totalToday === 0 ? 0 : Math.round((blockedToday / totalToday) * 100);

  const latestDecisions = useMemo(() => decisions.slice(0, 10), [decisions]);

  const decisionClass = (decision?: FraudDecision) => {
    if (decision === "ALLOW") return "table-chip allow";
    if (decision === "BLOCK") return "table-chip block";
    return "table-chip hold";
  };

  const decisionIcon = (decision?: FraudDecision) => {
    if (decision === "ALLOW") return "pi pi-check";
    if (decision === "BLOCK") return "pi pi-ban";
    return "pi pi-clock";
  };

  const formatScore = (value?: number) => (value == null ? "-" : value.toFixed(3));
  const formatDate = (value?: string) => (value ? new Date(value).toLocaleString() : "-");

  useEffect(() => {
    loadDecisions();
  }, [loadDecisions]);

  useEffect(() => {
    if (!autoRefresh) {
      if (intervalRef.current) {
        window.clearInterval(intervalRef.current);
        intervalRef.current = undefined;
      }
      return;
    }

    if (!intervalRef.current) {
      intervalRef.current = window.setInterval(loadDecisions, 15000);
    }

    return () => {
      if (intervalRef.current) {
        window.clearInterval(intervalRef.current);
        intervalRef.current = undefined;
      }
    };
  }, [autoRefresh, loadDecisions]);

  return (
    <>
      <div className="surface-card" style={{ marginBottom: 16 }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            flexWrap: "wrap",
            gap: 12
          }}
        >
          <div>
            <h2 className="section-title">Fraud Operations Overview</h2>
            <p className="subtle">High-level risk posture for today.</p>
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: 12, flexWrap: "wrap" }}>
            <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <InputSwitch checked={autoRefresh} onChange={(e) => setAutoRefresh(Boolean(e.value))} />
              <span className="subtle">Auto refresh (15s)</span>
            </div>
            <Button label="Refresh" icon="pi pi-refresh" severity="secondary" onClick={loadDecisions} />
          </div>
        </div>
      </div>

      <div className="card-grid" style={{ marginBottom: 24 }}>
        <div className="metric-card">
          <div className="metric-label">Transactions Today</div>
          <div className="metric-value">{totalToday}</div>
          <div className="subtle">Decisions recorded</div>
        </div>
        <div className="metric-card">
          <div className="metric-label">Blocked</div>
          <div className="metric-value" style={{ color: "var(--sp-danger)" }}>
            {blockedToday}
          </div>
          <div className="subtle">High-risk actions</div>
        </div>
        <div className="metric-card">
          <div className="metric-label">Allowed</div>
          <div className="metric-value" style={{ color: "var(--sp-success)" }}>
            {allowedToday}
          </div>
          <div className="subtle">Cleared transactions</div>
        </div>
        <div className="metric-card">
          <div className="metric-label">Block Rate</div>
          <div className="metric-value">{blockRate}%</div>
          <div className="subtle">Share of decisions</div>
        </div>
      </div>

      <div className="surface-card">
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <h3 className="section-title">Latest Decisions</h3>
          <span className="subtle">Last updated {lastRefresh}</span>
        </div>
        {errorMessage ? (
          <div className="subtle" style={{ marginBottom: 12 }}>
            {errorMessage}
          </div>
        ) : null}
        <DataTable value={latestDecisions} responsiveLayout="scroll" stripedRows>
          <Column field="transactionId" header="Transaction" />
          <Column
            field="finalDecision"
            header="Decision"
            body={(row: DecisionRecord) => (
              <span className={decisionClass(row.finalDecision)}>
                <i className={decisionIcon(row.finalDecision)}></i>
                {row.finalDecision}
              </span>
            )}
          />
          <Column field="ruleScore" header="Rule Score" body={(row: DecisionRecord) => formatScore(row.ruleScore)} />
          <Column field="mlScore" header="ML Score" body={(row: DecisionRecord) => formatScore(row.mlScore)} />
          <Column field="decisionReason" header="Reason" />
          <Column
            field="createdAt"
            header="Decided At"
            body={(row: DecisionRecord) => formatDate(row.createdAt)}
          />
        </DataTable>
      </div>
    </>
  );
};

export default Dashboard;
