"use client";
import { useEffect, useState } from "react";
import { authApi } from "@/api/client";
import { useAuthStore } from "@/store/authStore";
import type { User } from "@/lib/types";
export function SessionProvider({ children }: { children: React.ReactNode }) {
  const [unavailable, setUnavailable] = useState(false);
  useEffect(() => {
    // Clear credentials left by the old localStorage-based implementation.
    localStorage.removeItem("auth-storage");
    let active = true;
    async function restore() {
      const startedWith = useAuthStore.getState().user;
      try {
        const { data } = await authApi.get<{ user: User }>("/session");
        if (active && useAuthStore.getState().user === startedWith) {
          useAuthStore.getState().setUser(data.user);
          setUnavailable(false);
        }
      } catch (error: unknown) {
        const status = (error as { response?: { status?: number } }).response
          ?.status;
        if (active && useAuthStore.getState().user === startedWith) {
          if (status === 401) {
            useAuthStore.getState().setUser(null);
            setUnavailable(false);
          } else setUnavailable(true);
        }
      }
    }
    void restore();
    const listener = () => {
      if (document.visibilityState === "visible") void restore();
    };
    document.addEventListener("visibilitychange", listener);
    return () => {
      active = false;
      document.removeEventListener("visibilitychange", listener);
    };
  }, []);
  return (
    <>
      {unavailable && (
        <div className="connection-banner" role="alert">
          Learnyst services are unavailable. Start the app with{" "}
          <code>npm run dev</code>, then{" "}
          <button onClick={() => window.location.reload()}>retry</button>.
        </div>
      )}
      {children}
    </>
  );
}
