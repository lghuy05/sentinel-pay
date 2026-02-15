import { useEffect, useRef, useState } from "react";
import { Button } from "primereact/button";
import { InputSwitch } from "primereact/inputswitch";

import ServiceStatusCard from "../components/ServiceStatusCard";
import { fetchAllHealth, type ServiceHealth } from "../api/health";

const SystemStatus = () => {
  const [services, setServices] = useState<ServiceHealth[]>([]);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const intervalRef = useRef<number | undefined>(undefined);

  const loadHealth = async () => {
    const data = await fetchAllHealth();
    setServices(data);
  };

  useEffect(() => {
    loadHealth();
  }, []);

  useEffect(() => {
    if (!autoRefresh) {
      if (intervalRef.current) {
        window.clearInterval(intervalRef.current);
        intervalRef.current = undefined;
      }
      return;
    }

    if (!intervalRef.current) {
      intervalRef.current = window.setInterval(loadHealth, 15000);
    }

    return () => {
      if (intervalRef.current) {
        window.clearInterval(intervalRef.current);
        intervalRef.current = undefined;
      }
    };
  }, [autoRefresh]);

  return (
    <>
      <div className="surface-card" style={{ marginBottom: 16 }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            flexWrap: "wrap",
            gap: 12
          }}
        >
          <div>
            <h2 className="section-title">System Status</h2>
            <p className="subtle">Health check across each microservice.</p>
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: 12, flexWrap: "wrap" }}>
            <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <InputSwitch checked={autoRefresh} onChange={(e) => setAutoRefresh(Boolean(e.value))} />
              <span className="subtle">Auto refresh (15s)</span>
            </div>
            <Button label="Refresh" icon="pi pi-refresh" severity="secondary" onClick={loadHealth} />
          </div>
        </div>
      </div>

      <div className="card-grid">
        {services.map((service) => (
          <ServiceStatusCard key={service.name} service={service} />
        ))}
      </div>
    </>
  );
};

export default SystemStatus;
