import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || ""
});

export const fetchModelStatus = async () => {
  const { data } = await api.get("/ml/status");
  return data;
};

export const triggerRetrain = async () => {
  const { data } = await api.post("/ml/retrain");
  return data;
};

export const reloadModel = async () => {
  const { data } = await api.post("/ml/reload");
  return data;
};
