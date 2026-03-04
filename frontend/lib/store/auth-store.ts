import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { User } from '../types/auth';
import { authApi } from '../api/auth';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  refreshTimer: NodeJS.Timeout | null;

  // Actions
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, password: string, email?: string, url?: string, role?: string) => Promise<void>;
  logout: () => Promise<void>;
  logoutAll: () => Promise<void>;
  setUser: (user: User | null) => void;
  clearError: () => void;
  checkAuth: () => Promise<void>;
  startAutoRefresh: () => void;
  stopAutoRefresh: () => void;
  refreshSession: () => Promise<void>;
}

// Auto-refresh session every 14 minutes (session expires at 15min)
const REFRESH_INTERVAL = 14 * 60 * 1000;

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
      refreshTimer: null,

      startAutoRefresh: () => {
        const { stopAutoRefresh, refreshSession } = get();
        stopAutoRefresh(); // Clear any existing timer

        const timer = setInterval(async () => {
          try {
            await refreshSession();
          } catch (error) {
            console.error('Auto-refresh failed:', error);
            // If refresh fails, stop the timer and logout
            stopAutoRefresh();
            get().logout();
          }
        }, REFRESH_INTERVAL);

        set({ refreshTimer: timer });
      },

      stopAutoRefresh: () => {
        const { refreshTimer } = get();
        if (refreshTimer) {
          clearInterval(refreshTimer);
          set({ refreshTimer: null });
        }
      },

      refreshSession: async () => {
        try {
          await authApi.refresh();
          // Session cookie is automatically updated by the server
        } catch (error: any) {
          console.error('Session refresh failed:', error);
          throw error;
        }
      },

      login: async (username: string, password: string) => {
        set({ isLoading: true, error: null });
        try {
          await authApi.login({ username, password });
          // Get user after successful login
          const user = await authApi.getCurrentUser();
          set({
            user,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          });
          get().startAutoRefresh();
        } catch (error: any) {
          set({
            error: error.response?.data?.message || 'Login failed',
            isLoading: false,
            isAuthenticated: false,
            user: null,
          });
          throw error;
        }
      },

      register: async (username: string, password: string, email?: string, url?: string, role?: string) => {
        set({ isLoading: true, error: null });
        try {
          await authApi.register({ username, password, email, url, role });
          // After registration, need to login
          await get().login(username, password);
        } catch (error: any) {
          set({
            error: error.response?.data?.message || 'Registration failed',
            isLoading: false,
          });
          throw error;
        }
      },

      logout: async () => {
        const { stopAutoRefresh } = get();
        stopAutoRefresh();
        try {
          await authApi.logout();
        } catch (error) {
          console.error('Logout error:', error);
        } finally {
          set({
            user: null,
            isAuthenticated: false,
            error: null,
          });
        }
      },

      logoutAll: async () => {
        const { stopAutoRefresh } = get();
        stopAutoRefresh();
        try {
          await authApi.logoutAll();
        } catch (error) {
          console.error('Logout all error:', error);
        } finally {
          set({
            user: null,
            isAuthenticated: false,
            error: null,
          });
        }
      },

      setUser: (user: User | null) => {
        set({
          user,
          isAuthenticated: !!user,
        });
      },

      clearError: () => {
        set({ error: null });
      },

      checkAuth: async () => {
        set({ isLoading: true });
        try {
          const user = await authApi.getCurrentUser();
          set({
            user,
            isAuthenticated: true,
            isLoading: false,
          });
          // Start auto-refresh if user is authenticated
          get().startAutoRefresh();
        } catch (error) {
          set({
            user: null,
            isAuthenticated: false,
            isLoading: false,
          });
          get().stopAutoRefresh();
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
