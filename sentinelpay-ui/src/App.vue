<template>
  <div class="app-shell">
    <header class="navbar">
      <div class="navbar-inner">
        <div class="brand">
          <span class="brand-mark"></span>
          SentinelPay Ops Console
        </div>
        <nav class="nav-links">
          <RouterLink to="/">Dashboard</RouterLink>
          <RouterLink to="/accounts">Accounts</RouterLink>
          <RouterLink to="/transactions">Simulate Transaction</RouterLink>
          <RouterLink to="/decisions">Fraud Decisions</RouterLink>
          <RouterLink to="/status">System Status</RouterLink>
        </nav>
      </div>
    </header>

    <main class="app-main">
      <RouterView />
    </main>

    <Toast position="top-right" />
  </div>
</template>

<script setup lang="ts">
import Toast from "primevue/toast";
</script>
