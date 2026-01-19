<template>
  <div class="surface-card" v-if="decision">
    <h3 class="section-title">Decision Detail</h3>
    <div style="display: flex; gap: 16px; flex-wrap: wrap;">
      <div style="flex: 1; min-width: 220px;">
        <p class="subtle">Transaction</p>
        <h3>{{ decision.transactionId }}</h3>
        <p class="subtle">Decision</p>
        <div :class="decisionClass(decision.finalDecision)">
          <i :class="decisionIcon(decision.finalDecision)"></i>
          {{ decision.finalDecision }}
        </div>
      </div>
      <div style="flex: 1; min-width: 220px;">
        <p class="subtle">Score Breakdown</p>
        <ul style="list-style: none; padding: 0; margin: 0;">
          <li>Rule score: <strong>{{ formatScore(decision.ruleScore) }}</strong></li>
          <li>Rule band: <strong>{{ decision.ruleBand || "-" }}</strong></li>
          <li>ML score: <strong>{{ formatScore(decision.mlScore) }}</strong></li>
          <li>ML band: <strong>{{ decision.mlBand || "-" }}</strong></li>
          <li>Blacklist hit: <strong>{{ decision.blacklistHit ? "YES" : "NO" }}</strong></li>
        </ul>
      </div>
      <div style="flex: 1; min-width: 220px;">
        <p class="subtle">Explanation</p>
        <p>{{ explanationText }}</p>
        <p class="subtle">Decided At</p>
        <strong>{{ formatDate(decision.createdAt) }}</strong>
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
        <p class="subtle">Decision Reason</p>
        <p>{{ decision.decisionReason || "-" }}</p>
      </div>
      <div>
        <p class="subtle">Features</p>
        <pre style="max-height: 180px; overflow: auto;">{{ featureSnapshot }}</pre>
      </div>
      <div>
        <p class="subtle">Model Info</p>
        <p>Version: {{ decision.modelVersion || "-" }}</p>
        <p>Rule version: {{ decision.ruleVersion ?? "-" }}</p>
      </div>
    </div>
  </div>
  <div v-else class="surface-card">
    <h3 class="section-title">Decision Detail</h3>
    <p class="subtle">Select a decision to inspect the score breakdown and matched signals.</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { DecisionRecord, FraudDecision } from "../api/decisions";

const props = defineProps<{ decision: DecisionRecord | null }>();

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

const explanationText = computed(() => {
  if (!props.decision) return "";
  if (props.decision.finalDecision === "BLOCK") {
    return props.decision.decisionReason
      ? `Blocked due to ${props.decision.decisionReason}.`
      : "This transaction exceeded the risk threshold. Review rule and ML signals before releasing funds.";
  }
  if (props.decision.finalDecision === "HOLD") {
    return props.decision.decisionReason
      ? `Held for review due to ${props.decision.decisionReason}.`
      : "Transaction flagged for manual review based on combined risk signals.";
  }
  return props.decision.decisionReason
    ? `Allowed due to ${props.decision.decisionReason}.`
    : "Transaction cleared through the risk stack. Monitor downstream alerts if needed.";
});

const featureSnapshot = computed(() => {
  if (!props.decision?.featuresJson) return "{}";
  try {
    return JSON.stringify(JSON.parse(props.decision.featuresJson), null, 2);
  } catch {
    return props.decision.featuresJson || "{}";
  }
});
</script>
