<template>
  <div class="surface-card" style="margin-bottom: 16px;">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px;">
      <div>
        <h2 class="section-title">System Status</h2>
        <p class="subtle">Live health snapshots across the SentinelPay stack.</p>
      </div>
      <div style="display: flex; align-items: center; gap: 12px;">
        <div style="display: flex; align-items: center; gap: 8px;">
          <InputSwitch v-model="autoRefresh" />
          <span class="subtle">Auto refresh (15s)</span>
        </div>
        <Button label="Refresh" icon="pi pi-sync" severity="secondary" @click="loadHealth" />
      </div>
    </div>
  </div>

  <div class="card-grid">
    <ServiceStatusCard v-for="service in services" :key="service.name" :service="service" />
  </div>

  <div class="surface-card" style="margin-top: 16px;">
    <h3 class="section-title">Configured Endpoints</h3>
    <ul style="margin: 0; padding-left: 16px; color: var(--sp-muted);">
      <li v-for="service in catalog" :key="service.name">
        {{ service.name }} → {{ service.url || "(not set)" }}
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import Button from "primevue/button";
import InputSwitch from "primevue/inputswitch";

import ServiceStatusCard from "../components/ServiceStatusCard.vue";
import { fetchAllHealth, serviceCatalog, type ServiceHealth } from "../api/health";

const services = ref<ServiceHealth[]>([]);
const catalog = serviceCatalog;
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

const loadHealth = async () => {
  const scrollX = window.scrollX;
  const scrollY = window.scrollY;
  services.value = await fetchAllHealth();
  await nextTick();
  restoreScroll(scrollX, scrollY);
};

const startAutoRefresh = () => {
  if (intervalId) return;
  intervalId = window.setInterval(loadHealth, 15000);
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
  await loadHealth();
  if (autoRefresh.value) {
    startAutoRefresh();
  }
});

onBeforeUnmount(stopAutoRefresh);
</script>
