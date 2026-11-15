import { create } from "zustand";
import type { User } from "@/lib/types";
interface AuthState {
  user: User | null;
  ready: boolean;
  workspaceVersion: number;
  refreshWorkspace: () => void;
  setUser: (user: User | null) => void;
}
// The session credential is an HttpOnly cookie; only public profile data lives here.
export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  ready: false,
  workspaceVersion: 0,
  refreshWorkspace: () =>
    set((state) => ({ workspaceVersion: state.workspaceVersion + 1 })),
  setUser: (user) => set({ user, ready: true }),
}));
