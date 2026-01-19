<template>
  <DataTable
    :value="decisions"
    dataKey="transactionId"
    selectionMode="single"
    :selection="selected"
    @rowSelect="onRowSelect"
    :paginator="true"
    :rows="10"
    sortField="createdAt"
    :sortOrder="-1"
    stripedRows
    responsiveLayout="scroll"
  >
    <Column field="transactionId" header="Transaction" sortable />
    <Column field="finalDecision" header="Decision" sortable>
      <template #body="{ data }">
        <span :class="decisionClass(data.finalDecision)">
          <i :class="decisionIcon(data.finalDecision)"></i>
          {{ data.finalDecision }}
        </span>
      </template>
    </Column>
    <Column field="ruleScore" header="Rule" sortable>
      <template #body="{ data }">{{ formatScore(data.ruleScore) }}</template>
    </Column>
    <Column field="mlScore" header="ML" sortable>
      <template #body="{ data }">{{ formatScore(data.mlScore) }}</template>
    </Column>
    <Column field="ruleBand" header="Rule Band" sortable />
    <Column field="decisionReason" header="Reason" sortable />
    <Column field="createdAt" header="Decided" sortable>
      <template #body="{ data }">{{ formatDate(data.createdAt) }}</template>
    </Column>
  </DataTable>
</template>

<script setup lang="ts">
import DataTable from "primevue/datatable";
import Column from "primevue/column";
import type { DecisionRecord, FraudDecision } from "../api/decisions";

const props = defineProps<{
  decisions: DecisionRecord[];
  selected?: DecisionRecord | null;
}>();

const emit = defineEmits<{ (e: "select", decision: DecisionRecord): void }>();

const onRowSelect = (event: { data: DecisionRecord }) => {
  emit("select", event.data);
};

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
</script>
