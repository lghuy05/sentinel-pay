import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_ACCOUNTS_API_BASE_URL || import.meta.env.VITE_API_BASE_URL || ""
});

export type KycLevel = "BASIC" | "FULL";
export type AccountStatus = "ACTIVE" | "LOCKED";

export interface Account {
  userId: number;
  accountCountry: string;
  homeCurrency: string;
  createdAt: string;
  kycLevel: KycLevel;
  status: AccountStatus;
  balanceMinor: number;
}

export interface CreateAccountPayload {
  userId?: number;
  accountCountry: string;
  homeCurrency: string;
  createdAt?: string;
  kycLevel?: KycLevel;
  status?: AccountStatus;
  initialBalance?: number;
}

export interface UpdateAccountPayload {
  accountCountry?: string;
  createdAt?: string;
  kycLevel?: KycLevel;
  status?: AccountStatus;
}

export interface BalancePayload {
  amount: number;
  currency: string;
}

export const fetchAccounts = async (limit = 50, offset = 0) => {
  const { data } = await api.get<Account[]>("/api/v1/accounts", {
    params: { limit, offset }
  });
  return data;
};

export const fetchAccount = async (userId: number) => {
  const { data } = await api.get<Account>(`/api/v1/accounts/${userId}`);
  return data;
};

export const createAccount = async (payload: CreateAccountPayload) => {
  const { data } = await api.post<Account>("/api/v1/accounts", payload);
  return data;
};

export const updateAccount = async (userId: number, payload: UpdateAccountPayload) => {
  const { data } = await api.patch<Account>(`/api/v1/accounts/${userId}`, payload);
  return data;
};

export const topupAccount = async (userId: number, payload: BalancePayload) => {
  const { data } = await api.post<Account>(`/api/v1/accounts/${userId}/topup`, payload);
  return data;
};

export const deleteAccount = async (userId: number) => {
  await api.delete(`/api/v1/accounts/${userId}`);
};
