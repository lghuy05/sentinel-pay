import React from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import PrimeReact from "primereact/api";

import App from "./App";
import Accounts from "./views/Accounts";
import Dashboard from "./views/Dashboard";
import Decisions from "./views/Decisions";
import Feedback from "./views/Feedback";
import SystemStatus from "./views/SystemStatus";
import Transactions from "./views/Transactions";
import { ToastProvider } from "./components/ToastProvider";

import "primereact/resources/themes/lara-light-teal/theme.css";
import "primereact/resources/primereact.min.css";
import "primeicons/primeicons.css";
import "./styles/base.css";

PrimeReact.ripple = true;

const container = document.getElementById("app");

if (!container) {
  throw new Error("Root container #app not found");
}

createRoot(container).render(
  <React.StrictMode>
    <ToastProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<App />}>
            <Route index element={<Dashboard />} />
            <Route path="accounts" element={<Accounts />} />
            <Route path="transactions" element={<Transactions />} />
            <Route path="decisions" element={<Decisions />} />
            <Route path="feedback" element={<Feedback />} />
            <Route path="status" element={<SystemStatus />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ToastProvider>
  </React.StrictMode>
);
