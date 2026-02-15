import { useEffect, useMemo, useState } from "react";
import { Button } from "primereact/button";
import { Calendar } from "primereact/calendar";
import { Column } from "primereact/column";
import { DataTable } from "primereact/datatable";
import { Dialog } from "primereact/dialog";
import { Dropdown } from "primereact/dropdown";
import { InputNumber } from "primereact/inputnumber";

import {
  createAccount,
  deleteAccount,
  fetchAccounts,
  topupAccount,
  updateAccount,
  type Account,
  type AccountStatus,
  type CreateAccountPayload,
  type KycLevel
} from "../api/accounts";
import { useToast } from "../components/ToastProvider";

interface CreateFormState {
  userId: number | null;
  accountCountry: string;
  homeCurrency: string;
  createdAt: Date | null;
  kycLevel: KycLevel;
  status: AccountStatus;
  initialBalance: number;
}

interface EditFormState {
  accountCountry: string;
  createdAt: Date | null;
  kycLevel: KycLevel;
  status: AccountStatus;
}

interface TopupFormState {
  amount: number;
  currency: string;
}

const Accounts = () => {
  const { show } = useToast();
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [createVisible, setCreateVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [topupVisible, setTopupVisible] = useState(false);
  const [editingAccount, setEditingAccount] = useState<Account | null>(null);

  const countryOptions = useMemo(() => ["US", "VN", "SG", "JP", "CA", "MX"], []);
  const currencyOptions = useMemo(() => ["USD", "VND", "SGD", "JPY", "EUR"], []);
  const kycOptions = useMemo<KycLevel[]>(() => ["BASIC", "FULL"], []);
  const statusOptions = useMemo<AccountStatus[]>(() => ["ACTIVE", "LOCKED"], []);

  const [createForm, setCreateForm] = useState<CreateFormState>({
    userId: null,
    accountCountry: "US",
    homeCurrency: "USD",
    createdAt: new Date(),
    kycLevel: "FULL",
    status: "ACTIVE",
    initialBalance: 0
  });

  const [editForm, setEditForm] = useState<EditFormState>({
    accountCountry: "US",
    createdAt: new Date(),
    kycLevel: "FULL",
    status: "ACTIVE"
  });

  const [topupForm, setTopupForm] = useState<TopupFormState>({
    amount: 0,
    currency: "USD"
  });

  const loadAccounts = async () => {
    try {
      const data = await fetchAccounts(200, 0);
      setAccounts(data);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to load accounts";
      show({ severity: "warn", summary: "Accounts unavailable", detail: message, life: 3500 });
    }
  };

  const openCreate = () => {
    setCreateVisible(true);
  };

  const openEdit = (account: Account) => {
    setEditingAccount(account);
    setEditForm({
      accountCountry: account.accountCountry,
      createdAt: new Date(account.createdAt),
      kycLevel: account.kycLevel,
      status: account.status
    });
    setEditVisible(true);
  };

  const openTopup = (account: Account) => {
    setEditingAccount(account);
    setTopupForm({ amount: 0, currency: account.homeCurrency });
    setTopupVisible(true);
  };

  const submitCreate = async () => {
    const payload: CreateAccountPayload = {
      userId: createForm.userId ?? undefined,
      accountCountry: createForm.accountCountry,
      homeCurrency: createForm.homeCurrency,
      createdAt: createForm.createdAt?.toISOString(),
      kycLevel: createForm.kycLevel,
      status: createForm.status,
      initialBalance: createForm.initialBalance
    };
    try {
      const account = await createAccount(payload);
      show({ severity: "success", summary: "Account created", detail: `User ${account.userId}`, life: 3000 });
      setCreateVisible(false);
      await loadAccounts();
    } catch (error) {
      const message = error instanceof Error ? error.message : "Account creation failed";
      show({ severity: "error", summary: "Create failed", detail: message, life: 3500 });
    }
  };

  const submitEdit = async () => {
    if (!editingAccount) return;
    try {
      await updateAccount(editingAccount.userId, {
        accountCountry: editForm.accountCountry,
        createdAt: editForm.createdAt?.toISOString(),
        kycLevel: editForm.kycLevel,
        status: editForm.status
      });
      show({ severity: "success", summary: "Account updated", detail: `User ${editingAccount.userId}`, life: 3000 });
      setEditVisible(false);
      await loadAccounts();
    } catch (error) {
      const message = error instanceof Error ? error.message : "Account update failed";
      show({ severity: "error", summary: "Update failed", detail: message, life: 3500 });
    }
  };

  const submitTopup = async () => {
    if (!editingAccount) return;
    try {
      await topupAccount(editingAccount.userId, {
        amount: topupForm.amount,
        currency: topupForm.currency
      });
      show({ severity: "success", summary: "Balance updated", detail: `User ${editingAccount.userId}`, life: 3000 });
      setTopupVisible(false);
      await loadAccounts();
    } catch (error) {
      const message = error instanceof Error ? error.message : "Top-up failed";
      show({ severity: "error", summary: "Top-up failed", detail: message, life: 3500 });
    }
  };

  const createPreset = async (preset: "us-sender" | "vn-receiver") => {
    try {
      if (preset === "us-sender") {
        const createdAt = new Date();
        createdAt.setDate(createdAt.getDate() - 365);
        await createAccount({
          accountCountry: "US",
          homeCurrency: "USD",
          createdAt: createdAt.toISOString(),
          kycLevel: "FULL",
          status: "ACTIVE",
          initialBalance: 50000
        });
        show({ severity: "success", summary: "Preset created", detail: "US sender ready", life: 3000 });
      } else {
        const createdAt = new Date();
        createdAt.setDate(createdAt.getDate() - 2);
        await createAccount({
          accountCountry: "VN",
          homeCurrency: "VND",
          createdAt: createdAt.toISOString(),
          kycLevel: "BASIC",
          status: "ACTIVE",
          initialBalance: 0
        });
        show({ severity: "success", summary: "Preset created", detail: "VN receiver ready", life: 3000 });
      }
      await loadAccounts();
    } catch (error) {
      const message = error instanceof Error ? error.message : "Preset failed";
      show({ severity: "error", summary: "Preset failed", detail: message, life: 3500 });
    }
  };

  const confirmDelete = async (account: Account) => {
    const approved = window.confirm(`Delete account ${account.userId}? This cannot be undone.`);
    if (!approved) return;
    try {
      await deleteAccount(account.userId);
      show({ severity: "success", summary: "Account deleted", detail: `User ${account.userId}`, life: 3000 });
      await loadAccounts();
    } catch (error) {
      const message = error instanceof Error ? error.message : "Delete failed";
      show({ severity: "error", summary: "Delete failed", detail: message, life: 3500 });
    }
  };

  const formatDate = (value: string) => new Date(value).toLocaleString();
  const formatBalance = (account: Account) => `${account.balanceMinor.toLocaleString()} ${account.homeCurrency}`;

  useEffect(() => {
    loadAccounts();
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
            <h2 className="section-title">Accounts</h2>
            <p className="subtle">Create and manage test accounts for fraud simulations.</p>
          </div>
          <div style={{ display: "flex", gap: 12, flexWrap: "wrap" }}>
            <Button label="Refresh" icon="pi pi-refresh" severity="secondary" onClick={loadAccounts} />
            <Button label="New Account" icon="pi pi-plus" onClick={openCreate} />
          </div>
        </div>
      </div>

      <div className="surface-card" style={{ marginBottom: 16 }}>
        <h3 className="section-title">Presets</h3>
        <div style={{ display: "flex", gap: 12, flexWrap: "wrap" }}>
          <Button
            label="Create US sender (1y old, $50k)"
            icon="pi pi-user"
            severity="secondary"
            onClick={() => createPreset("us-sender")}
          />
          <Button
            label="Create VN receiver (2d old, 0 VND)"
            icon="pi pi-user"
            severity="secondary"
            onClick={() => createPreset("vn-receiver")}
          />
        </div>
      </div>

      <div className="surface-card">
        <DataTable
          value={accounts}
          dataKey="userId"
          paginator
          rows={10}
          sortField="userId"
          sortOrder={1}
          stripedRows
          responsiveLayout="scroll"
        >
          <Column field="userId" header="User ID" sortable />
          <Column field="accountCountry" header="Country" sortable />
          <Column field="homeCurrency" header="Currency" sortable />
          <Column
            field="createdAt"
            header="Created"
            sortable
            body={(row: Account) => formatDate(row.createdAt)}
          />
          <Column field="kycLevel" header="KYC" sortable />
          <Column field="status" header="Status" sortable />
          <Column
            field="balanceMinor"
            header="Balance"
            sortable
            body={(row: Account) => formatBalance(row)}
          />
          <Column
            header="Actions"
            body={(row: Account) => (
              <div style={{ display: "flex", gap: 8 }}>
                <Button label="Edit" icon="pi pi-pencil" severity="secondary" size="small" onClick={() => openEdit(row)} />
                <Button label="Top-up" icon="pi pi-wallet" size="small" onClick={() => openTopup(row)} />
                <Button
                  label="Delete"
                  icon="pi pi-trash"
                  severity="danger"
                  size="small"
                  onClick={() => confirmDelete(row)}
                />
              </div>
            )}
          />
        </DataTable>
      </div>

      <Dialog
        header="Create Account"
        visible={createVisible}
        onHide={() => setCreateVisible(false)}
        style={{ width: 420 }}
        modal
        footer={
          <>
            <Button label="Cancel" severity="secondary" onClick={() => setCreateVisible(false)} />
            <Button label="Create" icon="pi pi-check" onClick={submitCreate} />
          </>
        }
      >
        <div className="form-grid" style={{ gap: 16 }}>
          <div>
            <label className="subtle">User ID (optional)</label>
            <InputNumber
              value={createForm.userId}
              onValueChange={(e) =>
                setCreateForm((prev) => ({ ...prev, userId: e.value === null ? null : Number(e.value) }))
              }
              useGrouping={false}
              className="w-full"
            />
          </div>
          <div>
            <label className="subtle">Country</label>
            <Dropdown
              value={createForm.accountCountry}
              options={countryOptions}
              onChange={(e) => setCreateForm((prev) => ({ ...prev, accountCountry: e.value }))}
              className="w-full"
            />
          </div>
          <div>
            <label className="subtle">Home Currency</label>
            <Dropdown
              value={createForm.homeCurrency}
              options={currencyOptions}
              onChange={(e) => setCreateForm((prev) => ({ ...prev, homeCurrency: e.value }))}
              className="w-full"
            />
          </div>
          <div>
            <label className="subtle">Created At</label>
            <Calendar
              value={createForm.createdAt}
              onChange={(e) => setCreateForm((prev) => ({ ...prev, createdAt: e.value as Date }))}
              className="w-full"
              showTime
              hourFormat="24"
            />
          </div>
          <div>
            <label className="subtle">KYC Level</label>
            <Dropdown
              value={createForm.kycLevel}
              options={kycOptions}
              onChange={(e) => setCreateForm((prev) => ({ ...prev, kycLevel: e.value }))}
              className="w-full"
            />
          </div>
          <div>
            <label className="subtle">Status</label>
            <Dropdown
              value={createForm.status}
              options={statusOptions}
              onChange={(e) => setCreateForm((prev) => ({ ...prev, status: e.value }))}
              className="w-full"
            />
          </div>
          <div>
            <label className="subtle">Initial Balance</label>
            <InputNumber
              value={createForm.initialBalance}
              onValueChange={(e) => setCreateForm((prev) => ({ ...prev, initialBalance: Number(e.value || 0) }))}
              useGrouping
              min={0}
              className="w-full"
            />
          </div>
        </div>
      </Dialog>

      <Dialog
        header="Edit Account"
        visible={editVisible}
        onHide={() => setEditVisible(false)}
        style={{ width: 420 }}
        modal
        footer={
          <>
            <Button label="Cancel" severity="secondary" onClick={() => setEditVisible(false)} />
            <Button label="Save" icon="pi pi-check" onClick={submitEdit} />
          </>
        }
      >
        <div className="form-grid" style={{ gap: 16 }}>
          <div>
            <label className="subtle">Country</label>
            <Dropdown
              value={editForm.accountCountry}
              options={countryOptions}
              onChange={(e) => setEditForm((prev) => ({ ...prev, accountCountry: e.value }))}
              className="w-full"
            />
          </div>
          <div>
            <label className="subtle">Created At</label>
            <Calendar
              value={editForm.createdAt}
              onChange={(e) => setEditForm((prev) => ({ ...prev, createdAt: e.value as Date }))}
              className="w-full"
              showTime
              hourFormat="24"
            />
          </div>
          <div>
            <label className="subtle">KYC Level</label>
            <Dropdown
              value={editForm.kycLevel}
              options={kycOptions}
              onChange={(e) => setEditForm((prev) => ({ ...prev, kycLevel: e.value }))}
              className="w-full"
            />
          </div>
          <div>
            <label className="subtle">Status</label>
            <Dropdown
              value={editForm.status}
              options={statusOptions}
              onChange={(e) => setEditForm((prev) => ({ ...prev, status: e.value }))}
              className="w-full"
            />
          </div>
        </div>
      </Dialog>

      <Dialog
        header="Top-up Balance"
        visible={topupVisible}
        onHide={() => setTopupVisible(false)}
        style={{ width: 380 }}
        modal
        footer={
          <>
            <Button label="Cancel" severity="secondary" onClick={() => setTopupVisible(false)} />
            <Button label="Top-up" icon="pi pi-check" onClick={submitTopup} />
          </>
        }
      >
        <div className="form-grid" style={{ gap: 16 }}>
          <div>
            <label className="subtle">Amount</label>
            <InputNumber
              value={topupForm.amount}
              onValueChange={(e) => setTopupForm((prev) => ({ ...prev, amount: Number(e.value || 0) }))}
              useGrouping
              min={0}
              className="w-full"
            />
          </div>
          <div>
            <label className="subtle">Currency</label>
            <Dropdown
              value={topupForm.currency}
              options={currencyOptions}
              onChange={(e) => setTopupForm((prev) => ({ ...prev, currency: e.value }))}
              className="w-full"
            />
          </div>
        </div>
      </Dialog>
    </>
  );
};

export default Accounts;
