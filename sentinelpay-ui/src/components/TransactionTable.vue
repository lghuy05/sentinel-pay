<template>
  <DataTable
    :value="transactions"
    dataKey="transactionId"
    :paginator="true"
    :rows="10"
    sortField="receivedAt"
    :sortOrder="-1"
    stripedRows
    responsiveLayout="scroll"
  >
    <Column field="transactionId" header="Transaction" sortable />
    <Column field="type" header="Type" sortable />
    <Column field="senderUserId" header="Sender" sortable />
    <Column field="receiverUserId" header="Receiver" sortable />
    <Column field="merchantId" header="Merchant" sortable />
    <Column field="amount" header="Amount" sortable>
      <template #body="{ data }">{{ formatAmount(data.amount) }} {{ data.currency }}</template>
    </Column>
    <Column field="deviceId" header="Device" sortable />
    <Column field="eventTime" header="Event Time" sortable>
      <template #body="{ data }">{{ formatDate(data.eventTime) }}</template>
    </Column>
    <Column field="receivedAt" header="Received" sortable>
      <template #body="{ data }">{{ formatDate(data.receivedAt) }}</template>
    </Column>
  </DataTable>
</template>

<script setup lang="ts">
import DataTable from "primevue/datatable";
import Column from "primevue/column";
import type { TransactionResponse } from "../api/transactions";

defineProps<{ transactions: TransactionResponse[] }>();

const formatDate = (value: string) => new Date(value).toLocaleString();
const formatAmount = (value: number) => value.toLocaleString();
</script>
