import { DataTable } from "primereact/datatable";
import { Column } from "primereact/column";
import type { DecisionRecord, FraudDecision } from "../api/decisions";

interface Props {
  decisions: DecisionRecord[];
  selected?: DecisionRecord | null;
  onSelect: (decision: DecisionRecord) => void;
}

const DecisionTable = ({ decisions, selected, onSelect }: Props) => {
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

  return (
    <DataTable
      value={decisions}
      dataKey="transactionId"
      selectionMode="single"
      selection={selected ?? undefined}
      onSelectionChange={(event) => {
        if (event.value) {
          onSelect(event.value as DecisionRecord);
        }
      }}
      paginator
      rows={10}
      sortField="createdAt"
      sortOrder={-1}
      stripedRows
      responsiveLayout="scroll"
    >
      <Column field="transactionId" header="Transaction" sortable />
      <Column
        field="finalDecision"
        header="Decision"
        sortable
        body={(row: DecisionRecord) => (
          <span className={decisionClass(row.finalDecision)}>
            <i className={decisionIcon(row.finalDecision)}></i>
            {row.finalDecision}
          </span>
        )}
      />
      <Column field="ruleScore" header="Rule" sortable body={(row: DecisionRecord) => formatScore(row.ruleScore)} />
      <Column field="mlScore" header="ML" sortable body={(row: DecisionRecord) => formatScore(row.mlScore)} />
      <Column field="ruleBand" header="Rule Band" sortable />
      <Column field="decisionReason" header="Reason" sortable />
      <Column field="createdAt" header="Decided" sortable body={(row: DecisionRecord) => formatDate(row.createdAt)} />
    </DataTable>
  );
};

export default DecisionTable;
