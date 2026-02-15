import { useEffect, useMemo, useState } from "react";
import axios from "axios";
import { Button } from "primereact/button";
import { Calendar } from "primereact/calendar";
import { Dropdown } from "primereact/dropdown";
import { InputNumber } from "primereact/inputnumber";
import { InputSwitch } from "primereact/inputswitch";
import { InputText } from "primereact/inputtext";

import { fetchAccounts, type Account } from "../api/accounts";
import { fetchDecision, type DecisionRecord } from "../api/decisions";
import { createTransaction, type CreateTransactionPayload } from "../api/transactions";
import { useToast } from "./ToastProvider";

interface FormState {
  transactionId: string;
  type: CreateTransactionPayload["type"];
  senderUserId: number | null;
  receiverUserId: number | null;
  merchantId: number | null;
  amount: number;
  currency: string;
  deviceId: string;
  timestamp: Date;
}

const TransactionForm = () => {
  const { show } = useToast();

  const typeOptions = useMemo(
    () => [
      { label: "P2P", value: "P2P_TRANSFER" },
      { label: "Merchant", value: "MERCHANT_PAYMENT" }
    ],
    []
  );

  const currencyOptions = useMemo(() => ["VND", "USD", "SGD", "EUR", "JPY"], []);

  const [nightMode, setNightMode] = useState(false);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [decision, setDecision] = useState<DecisionRecord | null>(null);
  const [decisionState, setDecisionState] = useState<"idle" | "pending" | "ready" | "error">("idle");
  const [decisionMessage, setDecisionMessage] = useState("");
  const [lastTransactionId, setLastTransactionId] = useState<string | null>(null);

  const buildNightTimestamp = () => {
    const date = new Date();
    date.setUTCHours(2, 30, 0, 0);
    return date;
  };

  const createId = () => {
    if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
      return `tx-${crypto.randomUUID().slice(0, 8)}`;
    }
    return `tx-${Math.random().toString(36).slice(2, 10)}`;
  };

  const [form, setForm] = useState<FormState>(() => ({
    transactionId: createId(),
    type: "P2P_TRANSFER",
    senderUserId: null,
    receiverUserId: null,
    merchantId: null,
    amount: 150000,
    currency: "VND",
    deviceId: "device-known-001",
    timestamp: new Date()
  }));

  const accountOptions = useMemo(
    () =>
      accounts.map((account) => ({
        label: `${account.userId} · ${account.accountCountry} · ${ageDays(account.createdAt)}d · ${account.status}`,
        value: account.userId
      })),
    [accounts]
  );

  const selectedSender = useMemo(
    () => accounts.find((account) => account.userId === form.senderUserId) || null,
    [accounts, form.senderUserId]
  );

  const selectedReceiver = useMemo(
    () => accounts.find((account) => account.userId === form.receiverUserId) || null,
    [accounts, form.receiverUserId]
  );

  const crossBorderLabel = useMemo(() => {
    if (!selectedSender || !selectedReceiver) {
      return "";
    }
    return selectedSender.accountCountry !== selectedReceiver.accountCountry ? "Yes" : "No";
  }, [selectedSender, selectedReceiver]);

  const generateTransactionId = () => {
    setForm((prev) => ({ ...prev, transactionId: createId() }));
  };

  const applyNightMode = (enabled: boolean) => {
    setNightMode(enabled);
    setForm((prev) => ({ ...prev, timestamp: enabled ? buildNightTimestamp() : new Date() }));
  };

  const syncCurrency = (userId: number | null) => {
    const account = accounts.find((item) => item.userId === userId);
    if (account?.homeCurrency) {
      setForm((prev) => ({ ...prev, currency: account.homeCurrency }));
    }
  };

  const loadAccounts = async () => {
    try {
      const data = await fetchAccounts(200, 0);
      setAccounts(data);
      setForm((prev) => {
        const next = { ...prev };
        if (!prev.senderUserId && data.length) {
          next.senderUserId = data[0].userId;
          next.currency = data[0].homeCurrency;
        }
        if (!prev.receiverUserId && data.length > 1) {
          next.receiverUserId = data[1].userId;
        }
        return next;
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to load accounts";
      show({ severity: "warn", summary: "Accounts unavailable", detail: message, life: 3500 });
    }
  };

  const resetForm = () => {
    setNightMode(false);
    setForm((prev) => ({
      ...prev,
      transactionId: createId(),
      type: "P2P_TRANSFER",
      senderUserId: accounts[0]?.userId ?? null,
      receiverUserId: accounts[1]?.userId ?? null,
      merchantId: null,
      amount: 150000,
      currency: accounts[0]?.homeCurrency ?? "VND",
      deviceId: "device-known-001",
      timestamp: new Date()
    }));
    setDecision(null);
    setDecisionState("idle");
    setDecisionMessage("");
    setLastTransactionId(null);
  };

  const wait = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

  const pollDecision = async (transactionId: string) => {
    setDecisionState("pending");
    setDecisionMessage("");
    setLastTransactionId(transactionId);

    for (let attempt = 0; attempt < 12; attempt += 1) {
      try {
        const data = await fetchDecision(transactionId);
        setDecision(data);
        setDecisionState("ready");
        return;
      } catch (error) {
        if (axios.isAxiosError(error) && error.response?.status === 404) {
          await wait(1000);
          continue;
        }
        setDecisionState("error");
        setDecisionMessage("Decision fetch failed. Check fraud-orchestrator logs.");
        return;
      }
    }

    setDecisionState("error");
    setDecisionMessage("Decision not available yet. Try refreshing decisions.");
  };

  const submit = async () => {
    if (!form.senderUserId) {
      show({
        severity: "warn",
        summary: "Missing accounts",
        detail: "Select a sender account before sending.",
        life: 3500
      });
      return;
    }
    if (form.type === "P2P_TRANSFER" && !form.receiverUserId) {
      show({
        severity: "warn",
        summary: "Missing receiver",
        detail: "Select a receiver account for P2P transfers.",
        life: 3500
      });
      return;
    }
    if (form.type === "MERCHANT_PAYMENT" && !form.merchantId) {
      show({
        severity: "warn",
        summary: "Missing merchant",
        detail: "Enter a merchant ID for merchant payments.",
        life: 3500
      });
      return;
    }

    const transactionId = createId();
    setForm((prev) => ({ ...prev, transactionId }));

    const payload: CreateTransactionPayload = {
      transactionId,
      type: form.type,
      senderUserId: Number(form.senderUserId),
      receiverUserId: form.type === "P2P_TRANSFER" ? Number(form.receiverUserId) : null,
      merchantId: form.type === "MERCHANT_PAYMENT" ? Number(form.merchantId) : null,
      amount: Number(form.amount),
      currency: form.currency,
      deviceId: form.deviceId,
      timestamp: form.timestamp.toISOString()
    };

    try {
      await createTransaction(payload);
      show({
        severity: "success",
        summary: "Transaction accepted",
        detail: `Queued ${payload.transactionId} for ingestion`,
        life: 3500
      });
      generateTransactionId();
      await pollDecision(payload.transactionId);
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 409) {
        generateTransactionId();
        show({
          severity: "warn",
          summary: "Duplicate transaction ID",
          detail: "Generated a new transaction ID. Please retry.",
          life: 4000
        });
        return;
      }
      let message = "Failed to send transaction";
      if (axios.isAxiosError(error)) {
        const data = error.response?.data as { message?: string; error?: string } | string | undefined;
        if (typeof data === "string" && data.trim()) {
          message = data;
        } else if (data && typeof data === "object") {
          message = data.message || data.error || message;
        } else if (error.message) {
          message = error.message;
        }
      } else if (error instanceof Error) {
        message = error.message;
      }
      show({
        severity: "error",
        summary: "Transaction failed",
        detail: message,
        life: 4000
      });
    }
  };

  const decisionClass = (decisionValue: string) => {
    if (decisionValue === "ALLOW") return "table-chip allow";
    if (decisionValue === "BLOCK") return "table-chip block";
    return "table-chip hold";
  };

  const decisionIcon = (decisionValue: string) => {
    if (decisionValue === "ALLOW") return "pi pi-check";
    if (decisionValue === "BLOCK") return "pi pi-ban";
    return "pi pi-clock";
  };

  const formatScore = (value?: number) => (value == null ? "-" : value.toFixed(3));

  useEffect(() => {
    loadAccounts();
  }, []);

  return (
    <>
      <div className="surface-card">
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 16 }}>
          <div>
            <h2 className="section-title">Simulate Transaction</h2>
            <p className="subtle">Generate a transaction event and publish it into the pipeline.</p>
          </div>
          <Button label="New ID" icon="pi pi-refresh" severity="secondary" onClick={generateTransactionId} />
        </div>

        {!accounts.length ? (
          <div className="surface-card" style={{ marginBottom: 16 }}>
            <p className="subtle">No accounts available yet. Create accounts first in the Accounts page.</p>
          </div>
        ) : null}

        <div className="form-grid">
          <div>
            <label className="subtle">Transaction ID</label>
            <InputText
              value={form.transactionId}
              onChange={(e) => setForm((prev) => ({ ...prev, transactionId: e.target.value }))}
              className="w-full"
            />
          </div>

          <div>
            <label className="subtle">Type</label>
            <Dropdown
              value={form.type}
              options={typeOptions}
              optionLabel="label"
              optionValue="value"
              onChange={(e) =>
                setForm((prev) => ({
                  ...prev,
                  type: e.value,
                  receiverUserId: e.value === "P2P_TRANSFER" ? prev.receiverUserId : null,
                  merchantId: e.value === "MERCHANT_PAYMENT" ? prev.merchantId : null
                }))
              }
              className="w-full"
            />
          </div>

          <div>
            <label className="subtle">Sender Account</label>
            <Dropdown
              value={form.senderUserId}
              options={accountOptions}
              optionLabel="label"
              optionValue="value"
              filter
              onChange={(e) => {
                setForm((prev) => ({ ...prev, senderUserId: e.value }));
                syncCurrency(e.value);
              }}
              className="w-full"
            />
          </div>

          {form.type === "P2P_TRANSFER" ? (
            <div>
              <label className="subtle">Receiver Account</label>
              <Dropdown
                value={form.receiverUserId}
                options={accountOptions}
                optionLabel="label"
                optionValue="value"
                filter
                onChange={(e) => setForm((prev) => ({ ...prev, receiverUserId: e.value }))}
                className="w-full"
              />
            </div>
          ) : null}

          {form.type === "MERCHANT_PAYMENT" ? (
            <div>
              <label className="subtle">Merchant ID</label>
              <InputNumber
                value={form.merchantId}
                onValueChange={(e) => setForm((prev) => ({ ...prev, merchantId: e.value as number }))}
                useGrouping={false}
                className="w-full"
              />
            </div>
          ) : null}

          <div>
            <label className="subtle">Amount</label>
            <InputNumber
              value={form.amount}
              onValueChange={(e) => setForm((prev) => ({ ...prev, amount: Number(e.value || 0) }))}
              useGrouping
              min={0}
              className="w-full"
            />
          </div>

          <div>
            <label className="subtle">Currency</label>
            <Dropdown
              value={form.currency}
              options={currencyOptions}
              onChange={(e) => setForm((prev) => ({ ...prev, currency: e.value }))}
              className="w-full"
            />
          </div>

          <div>
            <label className="subtle">Device ID</label>
            <InputText
              value={form.deviceId}
              onChange={(e) => setForm((prev) => ({ ...prev, deviceId: e.target.value }))}
              className="w-full"
            />
          </div>

          <div>
            <label className="subtle">Timestamp</label>
            <Calendar
              value={form.timestamp}
              onChange={(e) => setForm((prev) => ({ ...prev, timestamp: e.value as Date }))}
              className="w-full"
              showTime
              hourFormat="24"
            />
          </div>
        </div>

        <div className="form-actions" style={{ justifyContent: "space-between" }}>
          <div style={{ display: "flex", alignItems: "center", gap: 12, flexWrap: "wrap" }}>
            <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <InputSwitch checked={nightMode} onChange={(e) => applyNightMode(Boolean(e.value))} />
              <span className="subtle">Night mode</span>
            </div>
            {crossBorderLabel ? (
              <span className="subtle">
                Cross-border: <strong>{crossBorderLabel}</strong>
              </span>
            ) : null}
          </div>
          <Button label="Refresh Accounts" icon="pi pi-sync" severity="secondary" onClick={loadAccounts} />
        </div>

        <div className="form-actions">
          <Button label="Send Transaction" icon="pi pi-send" onClick={submit} />
          <Button label="Reset" severity="secondary" icon="pi pi-undo" onClick={resetForm} />
        </div>
      </div>

      <div className="surface-card" style={{ marginTop: 16 }}>
        <h3 className="section-title">Decision Result</h3>
        {decisionState === "idle" ? (
          <div className="subtle">Send a transaction to see the decision.</div>
        ) : null}
        {decisionState === "pending" ? (
          <div className="subtle">Awaiting fraud decision for {lastTransactionId}…</div>
        ) : null}
        {decisionState === "error" ? <div className="subtle">{decisionMessage}</div> : null}
        {decisionState === "ready" && decision ? (
          <>
            <div style={{ display: "flex", flexWrap: "wrap", gap: 16 }}>
              <div>
                <p className="subtle">Decision</p>
                <div className={decisionClass(decision.finalDecision || "")}>
                  <i className={decisionIcon(decision.finalDecision || "")}></i>
                  {decision.finalDecision}
                </div>
              </div>
              <div>
                <p className="subtle">Decision Reason</p>
                <strong>{decision.decisionReason || "-"}</strong>
              </div>
              <div>
                <p className="subtle">Rule Score</p>
                <strong>{formatScore(decision.ruleScore)}</strong>
              </div>
              <div>
                <p className="subtle">ML Score</p>
                <strong>{formatScore(decision.mlScore)}</strong>
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
                <p className="subtle">Rule Band</p>
                <strong>{decision.ruleBand || "-"}</strong>
                <p className="subtle" style={{ marginTop: 8 }}>
                  ML Band
                </p>
                <strong>{decision.mlBand || "-"}</strong>
              </div>
              <div>
                <p className="subtle">Blacklist Hit</p>
                <strong>{decision.blacklistHit ? "YES" : "NO"}</strong>
              </div>
            </div>
          </>
        ) : null}
      </div>
    </>
  );
};

const ageDays = (createdAt: string) => {
  const created = new Date(createdAt).getTime();
  if (Number.isNaN(created)) {
    return 0;
  }
  return Math.max(0, Math.floor((Date.now() - created) / 86400000));
};

export default TransactionForm;
