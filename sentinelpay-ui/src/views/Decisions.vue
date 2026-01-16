<template>
  <div class="surface-card" style="margin-bottom: 16px;">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px;">
      <div>
        <h2 class="section-title">Fraud Decisions</h2>
        <p class="subtle">Latest decisions stored by alert-service.</p>
      </div>
      <div style="display: flex; align-items: center; gap: 12px; flex-wrap: wrap;">
        <div style="display: flex; align-items: center; gap: 8px;">
          <InputSwitch v-model="autoRefresh" />
          <span class="subtle">Auto refresh (10s)</span>
        </div>
        <span class="subtle">Last refresh: {{ lastRefresh }}</span>
        <Button label="Refresh" icon="pi pi-refresh" severity="secondary" @click="loadDecisions" />
      </div>
    </div>
  </div>

  <div v-if="loading" class="surface-card" style="display: flex; justify-content: center;">
    <ProgressSpinner />
  </div>

  <div v-else>
    <div v-if="errorMessage" class="surface-card" style="margin-bottom: 16px;">
      <p class="subtle">{{ errorMessage }}</p>
    </div>
    <div v-if="!decisions.length" class="surface-card">
      <p class="subtle">No decisions yet. Submit a transaction and wait for the pipeline.</p>
    </div>
    <div v-else style="display: grid; grid-template-columns: 1.6fr 1fr; gap: 16px;">
      <div class="surface-card">
        <DecisionTable :decisions="decisions" :selected="selected" @select="setSelected" />
      </div>
      <DecisionDetail :decision="selected" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import Button from "primevue/button";
import InputSwitch from "primevue/inputswitch";
import ProgressSpinner from "primevue/progressspinner";

import DecisionDetail from "../components/DecisionDetail.vue";
import DecisionTable from "../components/DecisionTable.vue";
import { fetchDecisions, type DecisionRecord } from "../api/decisions";

const decisions = ref<DecisionRecord[]>([]);
const selected = ref<DecisionRecord | null>(null);
const loading = ref(false);
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
  loading.value = true;
  try {
    const data = await fetchDecisions();
    decisions.value = data.sort(
      (a, b) => new Date(b.decidedAt).getTime() - new Date(a.decidedAt).getTime()
    );
    if (!selected.value && decisions.value.length) {
      selected.value = decisions.value[0];
    }
    errorMessage.value = "";
    lastRefresh.value = new Date().toLocaleTimeString();
  } catch (error) {
    const message = error instanceof Error ? error.message : "Failed to load decisions";
    errorMessage.value = message;
  } finally {
    loading.value = false;
    await nextTick();
    restoreScroll(scrollX, scrollY);
  }
};

const setSelected = (decision: DecisionRecord) => {
  selected.value = decision;
};

const startAutoRefresh = () => {
  if (intervalId) return;
  intervalId = window.setInterval(loadDecisions, 10000);
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

<style scoped>
@media (max-width: 1100px) {
  div[style*="grid-template-columns"] {
    grid-template-columns: 1fr !important;
  }
}
</style>
