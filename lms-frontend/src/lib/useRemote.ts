"use client";
import { useCallback, useEffect, useState } from "react";
import { useAuthStore } from "@/store/authStore";
import { coreApi, errorMessage } from "@/api/client";
export function useRemote<T>(url: string | null) {
  const workspaceVersion = useAuthStore((state) => state.workspaceVersion);
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [version, setVersion] = useState(0);
  const reload = useCallback(() => setVersion((value) => value + 1), []);
  useEffect(() => {
    if (!url) return;
    const controller = new AbortController();
    async function load() {
      setLoading(true);
      setError("");
      try {
        const response = await coreApi.get<T>(url!, {
          signal: controller.signal,
        });
        if (!controller.signal.aborted) setData(response.data);
      } catch (error) {
        if (!controller.signal.aborted) setError(errorMessage(error));
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => controller.abort();
  }, [url, version, workspaceVersion]);
  return { data, error, loading, reload };
}
