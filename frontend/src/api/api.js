import axios from "axios";

export const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

const client = axios.create({ baseURL: API_URL, headers: { "Content-Type": "application/json" } });

client.interceptors.request.use((config) => {
  try {
    const user = JSON.parse(localStorage.getItem("pulsepoll_user"));
    if (user?.token && user.token !== "demo-jwt-token") {
      config.headers.Authorization = `Bearer ${user.token}`;
    }
  } catch {}
  return config;
});

export const api = {
  signup: (data) => client.post("/api/auth/signup", data),
  login: (data) => client.post("/api/auth/login", data),
  createPoll: (data) => client.post("/api/polls", data),
  getPoll: (id) => client.get(`/api/polls/${id}`),
  getMyPolls: () => client.get("/api/polls"),
  vote: (id, optionId) => client.post(`/api/polls/${id}/vote`, { optionId })
};