<template>
  <div class="surface-card" style="margin-bottom: 16px;">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px;">
      <div>
        <h2 class="section-title">Fraud Operations Overview</h2>
        <p class="subtle">High-level risk posture for today.</p>
      </div>
      <div style="display: flex; align-items: center; gap: 12px; flex-wrap: wrap;">
        <div style="display: flex; align-items: center; gap: 8px;">
          <InputSwitch v-model="autoRefresh" />
          <span class="subtle">Auto refresh (15s)</span>
        </div>
        <Button label="Refresh" icon="pi pi-refresh" severity="secondary" @click="loadDecisions" />
      </div>
    </div>
  </div>

  <div class="card-grid" style="margin-bottom: 24px;">
    <div class="metric-card">
      <div class="metric-label">Transactions Today</div>
      <div class="metric-value">{{ totalToday }}</div>
      <div class="subtle">Decisions recorded</div>
    </div>
    <div class="metric-card">
      <div class="metric-label">Blocked</div>
      <div class="metric-value" style="color: var(--sp-danger);">{{ blockedToday }}</div>
      <div class="subtle">High-risk actions</div>
    </div>
    <div class="metric-card">
      <div class="metric-label">Allowed</div>
      <div class="metric-value" style="color: var(--sp-success);">{{ allowedToday }}</div>
      <div class="subtle">Cleared transactions</div>
    </div>
    <div class="metric-card">
      <div class="metric-label">Block Rate</div>
      <div class="metric-value">{{ blockRate }}%</div>
      <div class="subtle">Share of decisions</div>
    </div>
  </div>

  <div class="surface-card">
    <div style="display: flex; justify-content: space-between; align-items: center;">
      <h3 class="section-title">Latest Decisions</h3>
      <span class="subtle">Last updated {{ lastRefresh }}</span>
    </div>
    <div v-if="errorMessage" class="subtle" style="margin-bottom: 12px;">
      {{ errorMessage }}
    </div>
    <DataTable :value="latestDecisions" responsiveLayout="scroll" stripedRows>
      <Column field="transactionId" header="Transaction" />
      <Column field="finalDecision" header="Decision">
        <template #body="{ data }">
          <span :class="decisionClass(data.finalDecision)">
            <i :class="decisionIcon(data.finalDecision)"></i>
            {{ data.finalDecision }}
          </span>
        </template>
      </Column>
      <Column field="ruleScore" header="Rule Score">
        <template #body="{ data }">{{ formatScore(data.ruleScore) }}</template>
      </Column>
      <Column field="mlScore" header="ML Score">
        <template #body="{ data }">{{ formatScore(data.mlScore) }}</template>
      </Column>
      <Column field="decisionReason" header="Reason" />
      <Column field="createdAt" header="Decided At">
        <template #body="{ data }">{{ formatDate(data.createdAt) }}</template>
      </Column>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import Button from "primevue/button";
import DataTable from "primevue/datatable";
import Column from "primevue/column";
import InputSwitch from "primevue/inputswitch";

import { fetchDecisions, type DecisionRecord, type FraudDecision } from "../api/decisions";

const decisions = ref<DecisionRecord[]>([]);
const lastRefresh = ref("-");
const errorMessage = ref("");
const autoRefresh = ref(true);
let intervalId: number | undefined;

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
  try {
    const data = await fetchDecisions();
    decisions.value = data.sort(
      (a, b) => new Date(b.createdAt || 0).getTime() - new Date(a.createdAt || 0).getTime()
    );
    errorMessage.value = "";
  } catch (error) {
    const message = error instanceof Error ? error.message : "Failed to load decisions";
    errorMessage.value = message;
  } finally {
    lastRefresh.value = new Date().toLocaleTimeString();
    await nextTick();
    restoreScroll(scrollX, scrollY);
  }
};

const todayStart = new Date();

todayStart.setHours(0, 0, 0, 0);

const decisionsToday = computed(() =>
  decisions.value.filter((decision) => decision.createdAt && new Date(decision.createdAt) >= todayStart)
);

const totalToday = computed(() => decisionsToday.value.length);
const blockedToday = computed(() => decisionsToday.value.filter((d) => d.finalDecision === "BLOCK").length);
const allowedToday = computed(() => decisionsToday.value.filter((d) => d.finalDecision === "ALLOW").length);
const blockRate = computed(() =>
  totalToday.value === 0 ? 0 : Math.round((blockedToday.value / totalToday.value) * 100)
);

const latestDecisions = computed(() => decisions.value.slice(0, 10));

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

const startAutoRefresh = () => {
  if (intervalId) return;
  intervalId = window.setInterval(loadDecisions, 15000);
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
  await loadDecisions();
  if (autoRefresh.value) {
    startAutoRefresh();
  }
});

onBeforeUnmount(stopAutoRefresh);
</script>
