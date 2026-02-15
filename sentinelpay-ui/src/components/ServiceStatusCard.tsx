import { Card } from "primereact/card";
import { Tag } from "primereact/tag";
import type { ServiceHealth } from "../api/health";

const ServiceStatusCard = ({ service }: { service: ServiceHealth }) => {
  const badgeText = service.status === "up" ? "Connected" : service.status === "down" ? "Disconnected" : "Unknown";

  const tagSeverity = service.status === "up" ? "success" : service.status === "down" ? "danger" : "warning";

  const lastChecked = service.lastChecked ? new Date(service.lastChecked).toLocaleTimeString() : "-";

  return (
    <Card className="status-card">
      <div className="status-header">
        <span className="status-title">{service.name}</span>
        <Tag severity={tagSeverity} value={badgeText} />
      </div>
      {service.url ? <p className="status-url">{service.url}</p> : <p className="status-url">No endpoint configured</p>}
      <div className="status-meta">
        <span className="subtle">Last check: {lastChecked}</span>
        {service.details ? <span className="subtle">{service.details}</span> : null}
      </div>
    </Card>
  );
};

export default ServiceStatusCard;
