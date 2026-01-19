import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || ""
});

export const fetchKafkaStatus = async () => {
  const { data } = await api.get("/system/kafka");
  return data;
};

export const fetchRedisStatus = async () => {
  const { data } = await api.get("/system/redis");
  return data;
};

export const fetchServiceStatus = async () => {
  const { data } = await api.get("/system/services");
  return data;
};
