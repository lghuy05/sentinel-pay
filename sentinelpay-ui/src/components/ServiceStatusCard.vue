<template>
  <Card class="status-card">
    <template #content>
      <div class="status-header">
        <span class="status-title">{{ service.name }}</span>
        <Tag :severity="tagSeverity" :value="badgeText" />
      </div>
      <p class="status-url" v-if="service.url">{{ service.url }}</p>
      <p class="status-url" v-else>No endpoint configured</p>
      <div class="status-meta">
        <span class="subtle">Last check: {{ lastChecked }}</span>
        <span class="subtle" v-if="service.details">{{ service.details }}</span>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import Card from "primevue/card";
import Tag from "primevue/tag";
import type { ServiceHealth } from "../api/health";

const props = defineProps<{ service: ServiceHealth }>();

const badgeText = props.service.status === "up"
  ? "Connected"
  : props.service.status === "down"
    ? "Disconnected"
    : "Unknown";

const tagSeverity = props.service.status === "up"
  ? "success"
  : props.service.status === "down"
    ? "danger"
    : "warning";

const lastChecked = props.service.lastChecked
  ? new Date(props.service.lastChecked).toLocaleTimeString()
  : "-";
</script>
