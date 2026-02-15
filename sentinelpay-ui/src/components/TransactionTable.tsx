import { DataTable } from "primereact/datatable";
import { Column } from "primereact/column";
import type { TransactionResponse } from "../api/transactions";

const TransactionTable = ({ transactions }: { transactions: TransactionResponse[] }) => {
  const formatDate = (value: string) => new Date(value).toLocaleString();
  const formatAmount = (value: number) => value.toLocaleString();

  return (
    <DataTable
      value={transactions}
      dataKey="transactionId"
      paginator
      rows={10}
      sortField="receivedAt"
      sortOrder={-1}
      stripedRows
      responsiveLayout="scroll"
    >
      <Column field="transactionId" header="Transaction" sortable />
      <Column field="type" header="Type" sortable />
      <Column field="senderUserId" header="Sender" sortable />
      <Column field="receiverUserId" header="Receiver" sortable />
      <Column field="merchantId" header="Merchant" sortable />
      <Column
        field="amount"
        header="Amount"
        sortable
        body={(row: TransactionResponse) => `${formatAmount(row.amount)} ${row.currency}`}
      />
      <Column field="deviceId" header="Device" sortable />
      <Column field="eventTime" header="Event Time" sortable body={(row: TransactionResponse) => formatDate(row.eventTime)} />
      <Column field="receivedAt" header="Received" sortable body={(row: TransactionResponse) => formatDate(row.receivedAt)} />
    </DataTable>
  );
};

export default TransactionTable;
