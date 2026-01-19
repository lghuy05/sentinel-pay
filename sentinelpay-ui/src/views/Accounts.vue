<template>
  <div class="surface-card" style="margin-bottom: 16px;">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px;">
      <div>
        <h2 class="section-title">Accounts</h2>
        <p class="subtle">Create and manage test accounts for fraud simulations.</p>
      </div>
      <div style="display: flex; gap: 12px; flex-wrap: wrap;">
        <Button label="Refresh" icon="pi pi-refresh" severity="secondary" @click="loadAccounts" />
        <Button label="New Account" icon="pi pi-plus" @click="openCreate" />
      </div>
    </div>
  </div>

  <div class="surface-card" style="margin-bottom: 16px;">
    <h3 class="section-title">Presets</h3>
    <div style="display: flex; gap: 12px; flex-wrap: wrap;">
      <Button label="Create US sender (1y old, $50k)" icon="pi pi-user" severity="secondary" @click="createPreset('us-sender')" />
      <Button label="Create VN receiver (2d old, 0 VND)" icon="pi pi-user" severity="secondary" @click="createPreset('vn-receiver')" />
    </div>
  </div>

  <div class="surface-card">
    <DataTable
      :value="accounts"
      dataKey="userId"
      :paginator="true"
      :rows="10"
      sortField="userId"
      :sortOrder="1"
      stripedRows
      responsiveLayout="scroll"
    >
      <Column field="userId" header="User ID" sortable />
      <Column field="accountCountry" header="Country" sortable />
      <Column field="homeCurrency" header="Currency" sortable />
      <Column field="createdAt" header="Created" sortable>
        <template #body="{ data }">{{ formatDate(data.createdAt) }}</template>
      </Column>
      <Column field="kycLevel" header="KYC" sortable />
      <Column field="status" header="Status" sortable />
      <Column field="balanceMinor" header="Balance" sortable>
        <template #body="{ data }">{{ formatBalance(data) }}</template>
      </Column>
      <Column header="Actions">
        <template #body="{ data }">
          <div style="display: flex; gap: 8px;">
            <Button label="Edit" icon="pi pi-pencil" severity="secondary" size="small" @click="openEdit(data)" />
            <Button label="Top-up" icon="pi pi-wallet" size="small" @click="openTopup(data)" />
            <Button label="Delete" icon="pi pi-trash" severity="danger" size="small" @click="confirmDelete(data)" />
          </div>
        </template>
      </Column>
    </DataTable>
  </div>

  <Dialog v-model:visible="createVisible" header="Create Account" modal :style="{ width: '420px' }">
    <div class="form-grid" style="gap: 16px;">
      <div>
        <label class="subtle">User ID (optional)</label>
        <InputNumber v-model="createForm.userId" class="w-full" :useGrouping="false" />
      </div>
      <div>
        <label class="subtle">Country</label>
        <Dropdown v-model="createForm.accountCountry" :options="countryOptions" class="w-full" />
      </div>
      <div>
        <label class="subtle">Home Currency</label>
        <Dropdown v-model="createForm.homeCurrency" :options="currencyOptions" class="w-full" />
      </div>
      <div>
        <label class="subtle">Created At</label>
        <Calendar v-model="createForm.createdAt" class="w-full" showTime hourFormat="24" />
      </div>
      <div>
        <label class="subtle">KYC Level</label>
        <Dropdown v-model="createForm.kycLevel" :options="kycOptions" class="w-full" />
      </div>
      <div>
        <label class="subtle">Status</label>
        <Dropdown v-model="createForm.status" :options="statusOptions" class="w-full" />
      </div>
      <div>
        <label class="subtle">Initial Balance</label>
        <InputNumber v-model="createForm.initialBalance" class="w-full" :useGrouping="true" :min="0" />
      </div>
    </div>
    <template #footer>
      <Button label="Cancel" severity="secondary" @click="createVisible = false" />
      <Button label="Create" icon="pi pi-check" @click="submitCreate" />
    </template>
  </Dialog>

  <Dialog v-model:visible="editVisible" header="Edit Account" modal :style="{ width: '420px' }">
    <div class="form-grid" style="gap: 16px;">
      <div>
        <label class="subtle">Country</label>
        <Dropdown v-model="editForm.accountCountry" :options="countryOptions" class="w-full" />
      </div>
      <div>
        <label class="subtle">Created At</label>
        <Calendar v-model="editForm.createdAt" class="w-full" showTime hourFormat="24" />
      </div>
      <div>
        <label class="subtle">KYC Level</label>
        <Dropdown v-model="editForm.kycLevel" :options="kycOptions" class="w-full" />
      </div>
      <div>
        <label class="subtle">Status</label>
        <Dropdown v-model="editForm.status" :options="statusOptions" class="w-full" />
      </div>
    </div>
    <template #footer>
      <Button label="Cancel" severity="secondary" @click="editVisible = false" />
      <Button label="Save" icon="pi pi-check" @click="submitEdit" />
    </template>
  </Dialog>

  <Dialog v-model:visible="topupVisible" header="Top-up Balance" modal :style="{ width: '380px' }">
    <div class="form-grid" style="gap: 16px;">
      <div>
        <label class="subtle">Amount</label>
        <InputNumber v-model="topupForm.amount" class="w-full" :useGrouping="true" :min="0" />
      </div>
      <div>
        <label class="subtle">Currency</label>
        <Dropdown v-model="topupForm.currency" :options="currencyOptions" class="w-full" />
      </div>
    </div>
    <template #footer>
      <Button label="Cancel" severity="secondary" @click="topupVisible = false" />
      <Button label="Top-up" icon="pi pi-check" @click="submitTopup" />
    </template>
  </Dialog>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import Button from "primevue/button";
import Calendar from "primevue/calendar";
import Column from "primevue/column";
import DataTable from "primevue/datatable";
import Dialog from "primevue/dialog";
import Dropdown from "primevue/dropdown";
import InputNumber from "primevue/inputnumber";
import { useToast } from "primevue/usetoast";

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

const toast = useToast();

const accounts = ref<Account[]>([]);
const createVisible = ref(false);
const editVisible = ref(false);
const topupVisible = ref(false);
const editingAccount = ref<Account | null>(null);

const countryOptions = ["US", "VN", "SG", "JP", "CA", "MX"];
const currencyOptions = ["USD", "VND", "SGD", "JPY", "EUR"];
const kycOptions: KycLevel[] = ["BASIC", "FULL"];
const statusOptions: AccountStatus[] = ["ACTIVE", "LOCKED"];

const createForm = reactive({
  userId: null as number | null,
  accountCountry: "US",
  homeCurrency: "USD",
  createdAt: new Date(),
  kycLevel: "FULL" as KycLevel,
  status: "ACTIVE" as AccountStatus,
  initialBalance: 0
});

const editForm = reactive({
  accountCountry: "US",
  createdAt: new Date(),
  kycLevel: "FULL" as KycLevel,
  status: "ACTIVE" as AccountStatus
});

const topupForm = reactive({
  amount: 0,
  currency: "USD"
});

const loadAccounts = async () => {
  try {
    accounts.value = await fetchAccounts(200, 0);
  } catch (error) {
    const message = error instanceof Error ? error.message : "Failed to load accounts";
    toast.add({ severity: "warn", summary: "Accounts unavailable", detail: message, life: 3500 });
  }
};

const openCreate = () => {
  createVisible.value = true;
};

const openEdit = (account: Account) => {
  editingAccount.value = account;
  editForm.accountCountry = account.accountCountry;
  editForm.createdAt = new Date(account.createdAt);
  editForm.kycLevel = account.kycLevel;
  editForm.status = account.status;
  editVisible.value = true;
};

const openTopup = (account: Account) => {
  editingAccount.value = account;
  topupForm.amount = 0;
  topupForm.currency = account.homeCurrency;
  topupVisible.value = true;
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
    toast.add({ severity: "success", summary: "Account created", detail: `User ${account.userId}`, life: 3000 });
    createVisible.value = false;
    await loadAccounts();
  } catch (error) {
    const message = error instanceof Error ? error.message : "Account creation failed";
    toast.add({ severity: "error", summary: "Create failed", detail: message, life: 3500 });
  }
};

const submitEdit = async () => {
  if (!editingAccount.value) return;
  try {
    await updateAccount(editingAccount.value.userId, {
      accountCountry: editForm.accountCountry,
      createdAt: editForm.createdAt?.toISOString(),
      kycLevel: editForm.kycLevel,
      status: editForm.status
    });
    toast.add({ severity: "success", summary: "Account updated", detail: `User ${editingAccount.value.userId}`, life: 3000 });
    editVisible.value = false;
    await loadAccounts();
  } catch (error) {
    const message = error instanceof Error ? error.message : "Account update failed";
    toast.add({ severity: "error", summary: "Update failed", detail: message, life: 3500 });
  }
};

const submitTopup = async () => {
  if (!editingAccount.value) return;
  try {
    await topupAccount(editingAccount.value.userId, {
      amount: topupForm.amount,
      currency: topupForm.currency
    });
    toast.add({ severity: "success", summary: "Balance updated", detail: `User ${editingAccount.value.userId}`, life: 3000 });
    topupVisible.value = false;
    await loadAccounts();
  } catch (error) {
    const message = error instanceof Error ? error.message : "Top-up failed";
    toast.add({ severity: "error", summary: "Top-up failed", detail: message, life: 3500 });
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
      toast.add({ severity: "success", summary: "Preset created", detail: "US sender ready", life: 3000 });
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
      toast.add({ severity: "success", summary: "Preset created", detail: "VN receiver ready", life: 3000 });
    }
    await loadAccounts();
  } catch (error) {
    const message = error instanceof Error ? error.message : "Preset failed";
    toast.add({ severity: "error", summary: "Preset failed", detail: message, life: 3500 });
  }
};

const confirmDelete = async (account: Account) => {
  const approved = window.confirm(`Delete account ${account.userId}? This cannot be undone.`);
  if (!approved) return;
  try {
    await deleteAccount(account.userId);
    toast.add({ severity: "success", summary: "Account deleted", detail: `User ${account.userId}`, life: 3000 });
    await loadAccounts();
  } catch (error) {
    const message = error instanceof Error ? error.message : "Delete failed";
    toast.add({ severity: "error", summary: "Delete failed", detail: message, life: 3500 });
  }
};

const formatDate = (value: string) => new Date(value).toLocaleString();
const formatBalance = (account: Account) => `${account.balanceMinor.toLocaleString()} ${account.homeCurrency}`;

onMounted(loadAccounts);
</script>

<style scoped>
.w-full {
  width: 100%;
}
</style>
