<template>
  <div class="surface-card" v-if="decision">
    <h3 class="section-title">Decision Detail</h3>
    <div style="display: flex; gap: 16px; flex-wrap: wrap;">
      <div style="flex: 1; min-width: 220px;">
        <p class="subtle">Transaction</p>
        <h3>{{ decision.transactionId }}</h3>
        <p class="subtle">Decision</p>
        <div :class="decisionClass(decision.decision)">
          <i :class="decisionIcon(decision.decision)"></i>
          {{ decision.decision }}
        </div>
      </div>
      <div style="flex: 1; min-width: 220px;">
        <p class="subtle">Score Breakdown</p>
        <ul style="list-style: none; padding: 0; margin: 0;">
          <li>Final score: <strong>{{ formatScore(decision.finalScore) }}</strong></li>
          <li>ML score: <strong>{{ formatScore(decision.mlScore) }}</strong></li>
          <li>Rule score: <strong>{{ formatScore(decision.ruleScore) }}</strong></li>
          <li>Blacklist score: <strong>{{ formatScore(decision.blacklistScore) }}</strong></li>
          <li v-if="decision.riskScore !== undefined">Risk score: <strong>{{ formatScore(decision.riskScore) }}</strong></li>
          <li v-if="decision.riskLevel">Risk level: <strong>{{ decision.riskLevel }}</strong></li>
        </ul>
      </div>
      <div style="flex: 1; min-width: 220px;">
        <p class="subtle">Explanation</p>
        <p>{{ explanationText }}</p>
        <p class="subtle">Decided At</p>
        <strong>{{ formatDate(decision.decidedAt) }}</strong>
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
        <p class="subtle">Blacklist Matches</p>
        <div v-if="decision.blacklistMatches?.length" style="display: flex; flex-wrap: wrap; gap: 8px;">
          <span v-for="match in decision.blacklistMatches" :key="match" class="badge danger">{{ match }}</span>
        </div>
        <p v-else class="subtle">No blacklist matches</p>
      </div>
      <div>
        <p class="subtle">Hard Stops</p>
        <div v-if="decision.hardStopMatches?.length" style="display: flex; flex-wrap: wrap; gap: 8px;">
          <span v-for="match in decision.hardStopMatches" :key="match" class="badge danger">{{ match }}</span>
        </div>
        <p v-else class="subtle">No hard stops</p>
        <p v-if="decision.hardStopDecision" class="subtle" style="margin-top: 8px;">
          Decision override: <strong>{{ decision.hardStopDecision }}</strong>
        </p>
      </div>
      <div>
        <p class="subtle">Triggered Rules</p>
        <div v-if="decision.triggeredRules?.length" style="display: flex; flex-wrap: wrap; gap: 8px;">
          <span v-for="rule in decision.triggeredRules" :key="rule" class="badge info">{{ rule }}</span>
        </div>
        <p v-else class="subtle">No triggered rules</p>
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

const decisionClass = (decision: FraudDecision) => {
  if (decision === "ALLOW") return "table-chip allow";
  if (decision === "BLOCK") return "table-chip block";
  return "table-chip hold";
};

const decisionIcon = (decision: FraudDecision) => {
  if (decision === "ALLOW") return "pi pi-check";
  if (decision === "BLOCK") return "pi pi-ban";
  return "pi pi-clock";
};

const formatScore = (value: number) => value.toFixed(3);
const formatDate = (value: string) => new Date(value).toLocaleString();

const explanationText = computed(() => {
  if (!props.decision) return "";
  if (props.decision.decision === "BLOCK") {
    return "This transaction exceeded the risk threshold. Review rule and ML signals before releasing funds.";
  }
  if (props.decision.decision === "HOLD") {
    return "Transaction flagged for manual review based on combined risk signals.";
  }
  return "Transaction cleared through the risk stack. Monitor downstream alerts if needed.";
});
</script>
