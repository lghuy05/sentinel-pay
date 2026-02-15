import { useEffect, useRef, useState } from "react";
import { Button } from "primereact/button";
import { InputSwitch } from "primereact/inputswitch";
import { ProgressSpinner } from "primereact/progressspinner";

import DecisionDetail from "../components/DecisionDetail";
import DecisionTable from "../components/DecisionTable";
import { fetchDecisions, type DecisionRecord } from "../api/decisions";

const Decisions = () => {
  const [decisions, setDecisions] = useState<DecisionRecord[]>([]);
  const [selected, setSelected] = useState<DecisionRecord | null>(null);
  const [loading, setLoading] = useState(false);
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

  const loadDecisions = async () => {
    const scrollX = window.scrollX;
    const scrollY = window.scrollY;
    setLoading(true);
    try {
      const data = await fetchDecisions();
      const sorted = [...data].sort(
        (a, b) => new Date(b.createdAt || 0).getTime() - new Date(a.createdAt || 0).getTime()
      );
      setDecisions(sorted);
      if (!selected && sorted.length) {
        setSelected(sorted[0]);
      }
      setErrorMessage("");
      setLastRefresh(new Date().toLocaleTimeString());
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to load decisions";
      setErrorMessage(message);
    } finally {
      setLoading(false);
      restoreScroll(scrollX, scrollY);
    }
  };

  useEffect(() => {
    loadDecisions();
  }, []);

  useEffect(() => {
    if (!autoRefresh) {
      if (intervalRef.current) {
        window.clearInterval(intervalRef.current);
        intervalRef.current = undefined;
      }
      return;
    }

    if (!intervalRef.current) {
      intervalRef.current = window.setInterval(loadDecisions, 10000);
    }

    return () => {
      if (intervalRef.current) {
        window.clearInterval(intervalRef.current);
        intervalRef.current = undefined;
      }
    };
  }, [autoRefresh]);

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
            <h2 className="section-title">Fraud Decisions</h2>
            <p className="subtle">Latest decisions stored by fraud-orchestrator.</p>
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: 12, flexWrap: "wrap" }}>
            <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <InputSwitch checked={autoRefresh} onChange={(e) => setAutoRefresh(Boolean(e.value))} />
              <span className="subtle">Auto refresh (10s)</span>
            </div>
            <span className="subtle">Last refresh: {lastRefresh}</span>
            <Button label="Refresh" icon="pi pi-refresh" severity="secondary" onClick={loadDecisions} />
          </div>
        </div>
      </div>

      {loading ? (
        <div className="surface-card" style={{ display: "flex", justifyContent: "center" }}>
          <ProgressSpinner />
        </div>
      ) : (
        <>
          {errorMessage ? (
            <div className="surface-card" style={{ marginBottom: 16 }}>
              <p className="subtle">{errorMessage}</p>
            </div>
          ) : null}
          {!decisions.length ? (
            <div className="surface-card">
              <p className="subtle">No decisions yet. Submit a transaction and wait for the pipeline.</p>
            </div>
          ) : (
            <div className="decision-grid">
              <div className="surface-card">
                <DecisionTable decisions={decisions} selected={selected} onSelect={setSelected} />
              </div>
              <DecisionDetail decision={selected} />
            </div>
          )}
        </>
      )}
    </>
  );
};

export default Decisions;
