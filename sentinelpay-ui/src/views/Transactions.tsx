import { useEffect, useRef, useState } from "react";
import { Button } from "primereact/button";
import { InputSwitch } from "primereact/inputswitch";
import { ProgressSpinner } from "primereact/progressspinner";

import TransactionForm from "../components/TransactionForm";
import TransactionTable from "../components/TransactionTable";
import { fetchTransactions, type TransactionResponse } from "../api/transactions";

const Transactions = () => {
  const [transactions, setTransactions] = useState<TransactionResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [lastRefresh, setLastRefresh] = useState("-");
  const [autoRefresh, setAutoRefresh] = useState(true);
  const intervalRef = useRef<number | undefined>(undefined);

  const loadTransactions = async () => {
    setLoading(true);
    try {
      const data = await fetchTransactions(50);
      setTransactions(
        [...data].sort((a, b) => new Date(b.receivedAt).getTime() - new Date(a.receivedAt).getTime())
      );
      setLastRefresh(new Date().toLocaleTimeString());
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTransactions();
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
      intervalRef.current = window.setInterval(loadTransactions, 10000);
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
      <TransactionForm />

      <div className="surface-card" style={{ marginTop: 16 }}>
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
            <h2 className="section-title">Transaction History</h2>
            <p className="subtle">Recent ingested transactions stored by transaction-ingestor.</p>
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: 12, flexWrap: "wrap" }}>
            <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <InputSwitch checked={autoRefresh} onChange={(e) => setAutoRefresh(Boolean(e.value))} />
              <span className="subtle">Auto refresh (10s)</span>
            </div>
            <span className="subtle">Last refresh: {lastRefresh}</span>
            <Button label="Refresh" icon="pi pi-refresh" severity="secondary" onClick={loadTransactions} />
          </div>
        </div>
      </div>

      {loading ? (
        <div className="surface-card" style={{ marginTop: 16, display: "flex", justifyContent: "center" }}>
          <ProgressSpinner />
        </div>
      ) : (
        <div style={{ marginTop: 16 }}>
          {!transactions.length ? (
            <div className="surface-card">
              <p className="subtle">No transactions recorded yet. Submit a transaction above.</p>
            </div>
          ) : (
            <div className="surface-card">
              <TransactionTable transactions={transactions} />
            </div>
          )}
        </div>
      )}
    </>
  );
};

export default Transactions;
