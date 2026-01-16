<template>
  <TransactionForm />

  <div class="surface-card" style="margin-top: 16px;">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px;">
      <div>
        <h2 class="section-title">Transaction History</h2>
        <p class="subtle">Recent ingested transactions stored by transaction-ingestor.</p>
      </div>
      <div style="display: flex; align-items: center; gap: 12px; flex-wrap: wrap;">
        <div style="display: flex; align-items: center; gap: 8px;">
          <InputSwitch v-model="autoRefresh" />
          <span class="subtle">Auto refresh (10s)</span>
        </div>
        <span class="subtle">Last refresh: {{ lastRefresh }}</span>
        <Button label="Refresh" icon="pi pi-refresh" severity="secondary" @click="loadTransactions" />
      </div>
    </div>
  </div>

  <div v-if="loading" class="surface-card" style="margin-top: 16px; display: flex; justify-content: center;">
    <ProgressSpinner />
  </div>

  <div v-else style="margin-top: 16px;">
    <div v-if="!transactions.length" class="surface-card">
      <p class="subtle">No transactions recorded yet. Submit a transaction above.</p>
    </div>
    <div v-else class="surface-card">
      <TransactionTable :transactions="transactions" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import Button from "primevue/button";
import InputSwitch from "primevue/inputswitch";
import ProgressSpinner from "primevue/progressspinner";

import TransactionForm from "../components/TransactionForm.vue";
import TransactionTable from "../components/TransactionTable.vue";
import { fetchTransactions, type TransactionResponse } from "../api/transactions";

const transactions = ref<TransactionResponse[]>([]);
const loading = ref(false);
const lastRefresh = ref("-");
const autoRefresh = ref(true);
let intervalId: number | undefined;

const loadTransactions = async () => {
  loading.value = true;
  try {
    const data = await fetchTransactions(50);
    transactions.value = data.sort(
      (a, b) => new Date(b.receivedAt).getTime() - new Date(a.receivedAt).getTime()
    );
    lastRefresh.value = new Date().toLocaleTimeString();
  } finally {
    loading.value = false;
  }
};

const startAutoRefresh = () => {
  if (intervalId) return;
  intervalId = window.setInterval(loadTransactions, 10000);
};

const stopAutoRefresh = () => {
  if (intervalId) {
    window.clearInterval(intervalId);
    intervalId = undefined;
  }
};

watch(autoRefresh, (enabled) => {
  if (enabled) {
    startAutoRefresh();
  } else {
    stopAutoRefresh();
  }
});

onMounted(async () => {
  await loadTransactions();
  if (autoRefresh.value) {
    startAutoRefresh();
  }
});

onBeforeUnmount(stopAutoRefresh);
</script>
