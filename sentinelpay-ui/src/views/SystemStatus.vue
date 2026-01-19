<template>
  <div class="surface-card" style="margin-bottom: 16px;">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px;">
      <div>
        <h2 class="section-title">System Status</h2>
        <p class="subtle">Health check across each microservice.</p>
      </div>
      <div style="display: flex; align-items: center; gap: 12px; flex-wrap: wrap;">
        <div style="display: flex; align-items: center; gap: 8px;">
          <InputSwitch v-model="autoRefresh" />
          <span class="subtle">Auto refresh (15s)</span>
        </div>
        <Button label="Refresh" icon="pi pi-refresh" severity="secondary" @click="loadHealth" />
      </div>
    </div>
  </div>

  <div class="card-grid">
    <ServiceStatusCard v-for="service in services" :key="service.name" :service="service" />
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import Button from "primevue/button";
import InputSwitch from "primevue/inputswitch";

import ServiceStatusCard from "../components/ServiceStatusCard.vue";
import { fetchAllHealth, type ServiceHealth } from "../api/health";

const services = ref<ServiceHealth[]>([]);
const autoRefresh = ref(true);
let intervalId: number | undefined;

const loadHealth = async () => {
  services.value = await fetchAllHealth();
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
