<template>
  <div class="surface-card">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <div>
        <h2 class="section-title">Simulate Transaction</h2>
        <p class="subtle">Generate a transaction event and publish it into the pipeline.</p>
      </div>
      <Button label="New ID" icon="pi pi-refresh" severity="secondary" @click="generateTransactionId" />
    </div>

    <div v-if="!accounts.length" class="surface-card" style="margin-bottom: 16px;">
      <p class="subtle">No accounts available yet. Create accounts first in the Accounts page.</p>
    </div>

    <div class="form-grid">
      <div>
        <label class="subtle">Transaction ID</label>
        <InputText v-model="form.transactionId" class="w-full" />
      </div>

      <div>
        <label class="subtle">Type</label>
        <Dropdown
          v-model="form.type"
          :options="typeOptions"
          optionLabel="label"
          optionValue="value"
          class="w-full"
        />
      </div>

      <div>
        <label class="subtle">Sender Account</label>
        <Dropdown
          v-model="form.senderUserId"
          :options="accountOptions"
          optionLabel="label"
          optionValue="value"
          class="w-full"
          filter
          @change="syncCurrency"
        />
      </div>

      <div v-if="form.type === 'P2P_TRANSFER'">
        <label class="subtle">Receiver Account</label>
        <Dropdown
          v-model="form.receiverUserId"
          :options="accountOptions"
          optionLabel="label"
          optionValue="value"
          class="w-full"
          filter
        />
      </div>

      <div v-if="form.type === 'MERCHANT_PAYMENT'">
        <label class="subtle">Merchant ID</label>
        <InputNumber v-model="form.merchantId" class="w-full" :useGrouping="false" />
      </div>

      <div>
        <label class="subtle">Amount</label>
        <InputNumber v-model="form.amount" class="w-full" :useGrouping="true" />
      </div>

      <div>
        <label class="subtle">Currency</label>
        <Dropdown
          v-model="form.currency"
          :options="currencyOptions"
          class="w-full"
        />
      </div>

      <div>
        <label class="subtle">Device ID</label>
        <InputText v-model="form.deviceId" class="w-full" />
      </div>

      <div>
        <label class="subtle">Timestamp</label>
        <Calendar v-model="form.timestamp" class="w-full" showTime hourFormat="24" />
      </div>
    </div>

    <div class="form-actions" style="justify-content: space-between;">
      <div style="display: flex; align-items: center; gap: 12px; flex-wrap: wrap;">
        <div style="display: flex; align-items: center; gap: 8px;">
          <InputSwitch v-model="nightMode" @change="applyNightMode" />
          <span class="subtle">Night mode</span>
        </div>
        <span class="subtle" v-if="crossBorderLabel">
          Cross-border: <strong>{{ crossBorderLabel }}</strong>
        </span>
      </div>
      <Button label="Refresh Accounts" icon="pi pi-sync" severity="secondary" @click="loadAccounts" />
    </div>

    <div class="form-actions">
      <Button label="Send Transaction" icon="pi pi-send" @click="submit" />
      <Button label="Reset" severity="secondary" icon="pi pi-undo" @click="resetForm" />
    </div>
  </div>

  <div class="surface-card" style="margin-top: 16px;">
    <h3 class="section-title">Decision Result</h3>
    <div v-if="decisionState === 'idle'" class="subtle">Send a transaction to see the decision.</div>
    <div v-else-if="decisionState === 'pending'" class="subtle">
      Awaiting fraud decision for {{ lastTransactionId }}…
    </div>
    <div v-else-if="decisionState === 'error'" class="subtle">{{ decisionMessage }}</div>
    <div v-else-if="decision">
      <div style="display: flex; flex-wrap: wrap; gap: 16px;">
        <div>
          <p class="subtle">Decision</p>
          <div :class="decisionClass(decision.decision)">
            <i :class="decisionIcon(decision.decision)"></i>
            {{ decision.decision }}
          </div>
        </div>
        <div>
          <p class="subtle">Final Score</p>
          <strong>{{ formatScore(decision.finalScore) }}</strong>
        </div>
        <div v-if="decision.hardStopDecision">
          <p class="subtle">Hard Stop</p>
          <strong>{{ decision.hardStopDecision }}</strong>
        </div>
        <div v-if="decision.riskLevel">
          <p class="subtle">Risk Level</p>
          <strong>{{ decision.riskLevel }}</strong>
        </div>
      </div>
      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; margin-top: 16px;">
        <div>
          <p class="subtle">Rule Matches</p>
          <div v-if="decision.ruleMatches?.length" style="display: flex; flex-wrap: wrap; gap: 8px;">
            <span v-for="rule in decision.ruleMatches" :key="rule" class="badge warn">{{ rule }}</span>
          </div>
          <p v-else class="subtle">No rule matches</p>
        </div>
        <div>
          <p class="subtle">Hard Stops</p>
          <div v-if="decision.hardStopMatches?.length" style="display: flex; flex-wrap: wrap; gap: 8px;">
            <span v-for="match in decision.hardStopMatches" :key="match" class="badge danger">{{ match }}</span>
          </div>
          <p v-else class="subtle">No hard stops</p>
        </div>
        <div>
          <p class="subtle">Triggered Rules</p>
          <div v-if="decision.triggeredRules?.length" style="display: flex; flex-wrap: wrap; gap: 8px;">
            <span v-for="rule in decision.triggeredRules" :key="rule" class="badge info">{{ rule }}</span>
          </div>
          <p v-else class="subtle">No triggered rules</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import axios from "axios";
import Button from "primevue/button";
import Calendar from "primevue/calendar";
import Dropdown from "primevue/dropdown";
import InputNumber from "primevue/inputnumber";
import InputSwitch from "primevue/inputswitch";
import InputText from "primevue/inputtext";
import { useToast } from "primevue/usetoast";

import { fetchAccounts, type Account } from "../api/accounts";
import { fetchDecision, type DecisionRecord } from "../api/decisions";
import { createTransaction, type CreateTransactionPayload } from "../api/transactions";

const toast = useToast();

const typeOptions = [
  { label: "P2P", value: "P2P_TRANSFER" },
  { label: "Merchant", value: "MERCHANT_PAYMENT" }
];

const currencyOptions = ["VND", "USD", "SGD", "EUR", "JPY"];

const nightMode = ref(false);
const accounts = ref<Account[]>([]);
const decision = ref<DecisionRecord | null>(null);
const decisionState = ref<"idle" | "pending" | "ready" | "error">("idle");
const decisionMessage = ref("");
const lastTransactionId = ref<string | null>(null);

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

const form = reactive({
  transactionId: createId(),
  type: "P2P_TRANSFER",
  senderUserId: null as number | null,
  receiverUserId: null as number | null,
  merchantId: null as number | null,
  amount: 150000,
  currency: "VND",
  deviceId: "device-known-001",
  timestamp: new Date()
});

const generateTransactionId = () => {
  form.transactionId = createId();
};

const applyNightMode = () => {
  form.timestamp = nightMode.value ? buildNightTimestamp() : new Date();
};

const accountOptions = computed(() =>
  accounts.value.map((account) => ({
    label: `${account.userId} · ${account.accountCountry} · ${ageDays(account.createdAt)}d · ${account.status}`,
    value: account.userId
  }))
);

const selectedSender = computed(() =>
  accounts.value.find((account) => account.userId === form.senderUserId) || null
);

const selectedReceiver = computed(() =>
  accounts.value.find((account) => account.userId === form.receiverUserId) || null
);

const crossBorderLabel = computed(() => {
  if (!selectedSender.value || !selectedReceiver.value) {
    return "";
  }
  return selectedSender.value.accountCountry !== selectedReceiver.value.accountCountry ? "Yes" : "No";
});

const ageDays = (createdAt: string) => {
  const created = new Date(createdAt).getTime();
  if (Number.isNaN(created)) {
    return 0;
  }
  return Math.max(0, Math.floor((Date.now() - created) / 86400000));
};

const syncCurrency = () => {
  if (selectedSender.value?.homeCurrency) {
    form.currency = selectedSender.value.homeCurrency;
  }
};

const loadAccounts = async () => {
  try {
    accounts.value = await fetchAccounts(200, 0);
    if (!form.senderUserId && accounts.value.length) {
      form.senderUserId = accounts.value[0].userId;
      syncCurrency();
    }
    if (!form.receiverUserId && accounts.value.length > 1) {
      form.receiverUserId = accounts.value[1].userId;
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : "Failed to load accounts";
    toast.add({ severity: "warn", summary: "Accounts unavailable", detail: message, life: 3500 });
  }
};

const resetForm = () => {
  nightMode.value = false;
  form.transactionId = createId();
  form.type = "P2P_TRANSFER";
  form.senderUserId = accounts.value[0]?.userId ?? null;
  form.receiverUserId = accounts.value[1]?.userId ?? null;
  form.merchantId = null;
  form.amount = 150000;
  form.currency = selectedSender.value?.homeCurrency ?? "VND";
  form.deviceId = "device-known-001";
  form.timestamp = new Date();
  decision.value = null;
  decisionState.value = "idle";
  decisionMessage.value = "";
  lastTransactionId.value = null;
};

const wait = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

const pollDecision = async (transactionId: string) => {
  decisionState.value = "pending";
  decisionMessage.value = "";
  lastTransactionId.value = transactionId;

  for (let attempt = 0; attempt < 12; attempt += 1) {
    try {
      const data = await fetchDecision(transactionId);
      decision.value = data;
      decisionState.value = "ready";
      return;
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 404) {
        await wait(1000);
        continue;
      }
      decisionState.value = "error";
      decisionMessage.value = "Decision fetch failed. Check alert-service logs.";
      return;
    }
  }

  decisionState.value = "error";
  decisionMessage.value = "Decision not available yet. Try refreshing decisions.";
};

const submit = async () => {
  if (!form.senderUserId) {
    toast.add({
      severity: "warn",
      summary: "Missing accounts",
      detail: "Select a sender account before sending.",
      life: 3500
    });
    return;
  }
  if (form.type === "P2P_TRANSFER" && !form.receiverUserId) {
    toast.add({
      severity: "warn",
      summary: "Missing receiver",
      detail: "Select a receiver account for P2P transfers.",
      life: 3500
    });
    return;
  }
  if (form.type === "MERCHANT_PAYMENT" && !form.merchantId) {
    toast.add({
      severity: "warn",
      summary: "Missing merchant",
      detail: "Enter a merchant ID for merchant payments.",
      life: 3500
    });
    return;
  }
  // Always mint a fresh transaction ID per send.
  form.transactionId = createId();
  const payload: CreateTransactionPayload = {
    transactionId: form.transactionId,
    type: form.type as CreateTransactionPayload["type"],
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
    toast.add({
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
      toast.add({
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
    toast.add({
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

const formatScore = (value: number) => value.toFixed(3);

onMounted(loadAccounts);
</script>

<style scoped>
.w-full {
  width: 100%;
}
</style>
