import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_TRANSACTIONS_API_BASE_URL || import.meta.env.VITE_API_BASE_URL || ""
});

export type TransactionType = "P2P_TRANSFER" | "MERCHANT_PAYMENT";

export interface CreateTransactionPayload {
  transactionId: string;
  type: TransactionType;
  senderUserId: number;
  receiverUserId?: number | null;
  merchantId?: number | null;
  amount: number;
  currency: string;
  deviceId: string;
  timestamp: string;
}

export interface TransactionResponse {
  id: number;
  transactionId: string;
  type: TransactionType;
  senderUserId: number;
  receiverUserId?: number | null;
  merchantId?: number | null;
  amount: number;
  currency: string;
  deviceId: string;
  eventTime: string;
  receivedAt: string;
}

export const createTransaction = async (payload: CreateTransactionPayload) => {
  const { data } = await api.post<TransactionResponse>("/api/v1/transactions", payload);
  return data;
};

export const fetchTransactions = async (limit = 50) => {
  const { data } = await api.get<TransactionResponse[]>("/api/v1/transactions", {
    params: { limit }
  });
  return data;
};
