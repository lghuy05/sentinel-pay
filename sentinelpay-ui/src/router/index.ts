import { createRouter, createWebHistory } from "vue-router";

import Dashboard from "../views/Dashboard.vue";
import Accounts from "../views/Accounts.vue";
import Transactions from "../views/Transactions.vue";
import Decisions from "../views/Decisions.vue";
import SystemStatus from "../views/SystemStatus.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", name: "dashboard", component: Dashboard },
    { path: "/accounts", name: "accounts", component: Accounts },
    { path: "/transactions", name: "transactions", component: Transactions },
    { path: "/decisions", name: "decisions", component: Decisions },
    { path: "/status", name: "status", component: SystemStatus }
  ]
});

export default router;
