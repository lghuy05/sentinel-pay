<template>
  <div class="surface-card" style="margin-bottom: 16px;">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px;">
      <div>
        <h2 class="section-title">Feedback Queue</h2>
        <p class="subtle">Manually label ML or HOLD decisions for retraining.</p>
      </div>
      <div style="display: flex; align-items: center; gap: 12px; flex-wrap: wrap;">
        <Button label="Refresh" icon="pi pi-refresh" severity="secondary" @click="loadQueue" />
      </div>
    </div>
  </div>

  <div v-if="loading" class="surface-card" style="display: flex; justify-content: center;">
    <ProgressSpinner />
  </div>

  <div v-else>
    <div v-if="!queue.length" class="surface-card">
      <p class="subtle">Queue is empty.</p>
    </div>
    <div v-else class="surface-card">
      <DataTable :value="queue" responsiveLayout="scroll" stripedRows>
        <Column field="transactionId" header="Transaction" />
        <Column field="amount" header="Amount">
          <template #body="{ data }">{{ formatAmount(data.amount) }}</template>
        </Column>
        <Column field="ruleScore" header="Rule">
          <template #body="{ data }">{{ formatScore(data.ruleScore) }}</template>
        </Column>
        <Column field="mlScore" header="ML">
          <template #body="{ data }">{{ formatScore(data.mlScore) }}</template>
        </Column>
        <Column field="finalDecision" header="Decision">
          <template #body="{ data }">
            <span :class="decisionClass(data.finalDecision)">
              <i :class="decisionIcon(data.finalDecision)"></i>
              {{ data.finalDecision }}
            </span>
          </template>
        </Column>
        <Column field="decisionReason" header="Reason" />
        <Column header="Actions">
          <template #body="{ data }">
            <div style="display: flex; gap: 8px; flex-wrap: wrap;">
              <Button label="✅ Legit" size="small" @click="submit(data.transactionId, 0)" />
              <Button label="🚨 Fraud" severity="danger" size="small" @click="submit(data.transactionId, 1)" />
            </div>
          </template>
        </Column>
      </DataTable>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import Button from "primevue/button";
import Column from "primevue/column";
import DataTable from "primevue/datatable";
import ProgressSpinner from "primevue/progressspinner";
import { fetchUnreviewedDecisions, type DecisionRecord } from "../api/decisions";
import { submitFeedback } from "../api/feedback";

const queue = ref<DecisionRecord[]>([]);
const loading = ref(false);

const filterQueue = (records: DecisionRecord[]) =>
  records.filter(
    (record) =>
      record.finalDecision === "HOLD" ||
      (record.decisionReason || "").startsWith("ML_")
  );

const loadQueue = async () => {
  loading.value = true;
  try {
    const data = await fetchUnreviewedDecisions(120);
    queue.value = filterQueue(data);
  } finally {
    loading.value = false;
  }
};

const submit = async (transactionId: string, label: 0 | 1) => {
  await submitFeedback(transactionId, label);
  queue.value = queue.value.filter((item) => item.transactionId !== transactionId);
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

loadQueue();
</script>
