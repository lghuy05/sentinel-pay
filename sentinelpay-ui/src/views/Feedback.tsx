import { useEffect, useState } from "react";
import { Button } from "primereact/button";
import { Column } from "primereact/column";
import { DataTable } from "primereact/datatable";
import { ProgressSpinner } from "primereact/progressspinner";

import { fetchUnreviewedDecisions, type DecisionRecord } from "../api/decisions";
import { submitFeedback } from "../api/feedback";

const Feedback = () => {
  const [queue, setQueue] = useState<DecisionRecord[]>([]);
  const [loading, setLoading] = useState(false);

  const filterQueue = (records: DecisionRecord[]) =>
    records.filter(
      (record) => record.finalDecision === "HOLD" || (record.decisionReason || "").startsWith("ML_")
    );

  const loadQueue = async () => {
    setLoading(true);
    try {
      const data = await fetchUnreviewedDecisions(120);
      setQueue(filterQueue(data));
    } finally {
      setLoading(false);
    }
  };

  const submit = async (transactionId: string, label: 0 | 1) => {
    await submitFeedback(transactionId, label);
    setQueue((prev) => prev.filter((item) => item.transactionId !== transactionId));
  };

  const formatAmount = (amount?: number) => (amount == null ? "-" : amount.toLocaleString());
  const formatScore = (score?: number) => (score == null ? "-" : score.toFixed(2));
  const decisionClass = (decision?: string) => {
    if (decision === "ALLOW") return "table-chip allow";
    if (decision === "BLOCK") return "table-chip block";
    return "table-chip hold";
  };
  const decisionIcon = (decision?: string) => {
    if (decision === "ALLOW") return "pi pi-check";
    if (decision === "BLOCK") return "pi pi-ban";
    return "pi pi-clock";
  };

  useEffect(() => {
    loadQueue();
  }, []);

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
            <h2 className="section-title">Feedback Queue</h2>
            <p className="subtle">Manually label ML or HOLD decisions for retraining.</p>
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: 12, flexWrap: "wrap" }}>
            <Button label="Refresh" icon="pi pi-refresh" severity="secondary" onClick={loadQueue} />
          </div>
        </div>
      </div>

      {loading ? (
        <div className="surface-card" style={{ display: "flex", justifyContent: "center" }}>
          <ProgressSpinner />
        </div>
      ) : (
        <>
          {!queue.length ? (
            <div className="surface-card">
              <p className="subtle">Queue is empty.</p>
            </div>
          ) : (
            <div className="surface-card">
              <DataTable value={queue} responsiveLayout="scroll" stripedRows>
                <Column field="transactionId" header="Transaction" />
                <Column field="amount" header="Amount" body={(row: DecisionRecord) => formatAmount(row.amount)} />
                <Column field="ruleScore" header="Rule" body={(row: DecisionRecord) => formatScore(row.ruleScore)} />
                <Column field="mlScore" header="ML" body={(row: DecisionRecord) => formatScore(row.mlScore)} />
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
                <Column field="decisionReason" header="Reason" />
                <Column
                  header="Actions"
                  body={(row: DecisionRecord) => (
                    <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
                      <Button label=" Legit" size="small" onClick={() => submit(row.transactionId, 0)} />
                      <Button label=" Fraud" severity="danger" size="small" onClick={() => submit(row.transactionId, 1)} />
                    </div>
                  )}
                />
              </DataTable>
            </div>
          )}
        </>
      )}
    </>
  );
};

export default Feedback;
