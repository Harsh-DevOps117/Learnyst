import axios from "axios";
import { useAuthStore } from "@/store/authStore";
export const authApi = axios.create({
  baseURL: "/api/auth",
  withCredentials: true,
  timeout: 15000,
});
export const coreApi = axios.create({
  baseURL: "/api/core",
  withCredentials: true,
  timeout: 20000,
});
coreApi.interceptors.response.use(
  (response) => response,
  (error) => {
    if (axios.isAxiosError(error) && error.response?.status === 401)
      useAuthStore.getState().setUser(null);
    return Promise.reject(error);
  },
);
export function errorMessage(
  error: unknown,
  fallback = "Something went wrong. Please try again.",
) {
  if (axios.isAxiosError(error)) {
    if (!error.response)
      return "Could not connect to Learnyst. Check that the local services are running.";
    return error.response.data?.error || fallback;
  }
  return fallback;
}
