import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || ""
});

export const submitFeedback = async (transactionId: string, label: 0 | 1) => {
  await api.post("/feedback", { transactionId, label });
};
