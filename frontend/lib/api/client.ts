import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from 'axios';
import { ApiError } from '../types/auth';

// Create axios instance with base configuration
const apiClient: AxiosInstance = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:9520',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Send cookies with requests
});

let isRefreshing = false;
let failedQueue: any[] = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach(prom => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

// Request interceptor
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);

// Protected routes that require auth
const PROTECTED_ROUTES = ['/profile', '/settings', '/dashboard'];

const isProtectedRoute = () => {
  if (typeof window === 'undefined') return false;
  const path = window.location.pathname;
  return PROTECTED_ROUTES.some(route => path.startsWith(route));
};

// Response interceptor with automatic token refresh
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error: AxiosError<ApiError>) => {
    const originalRequest: any = error.config;

    // Handle 401 with auto-refresh (except for refresh endpoint)
    if (error.response?.status === 401 &&
      !originalRequest._retry &&
      !originalRequest.url?.includes('/auth/refresh')) {

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        }).then(() => {
          return apiClient(originalRequest);
        }).catch((err) => {
          return Promise.reject(err);
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        await apiClient.post('/auth/refresh');
        processQueue(null);
        return apiClient(originalRequest);
      } catch (refreshError) {
        processQueue(refreshError, null);

        // Clear storage on refresh failure
        if (typeof window !== 'undefined') {
          localStorage.removeItem('auth-storage');

          // Auto-redirect to login if on protected route
          if (isProtectedRoute()) {
            window.location.href = '/login';
          }
        }

        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    // Handle other errors
    if (error.response) {
      const status = error.response.status;
      const data = error.response.data;

      return Promise.reject({
        message: data?.message || data?.error || 'An error occurred',
        status,
        data,
      });
    } else if (error.request) {
      return Promise.reject({
        message: 'Network error. Please check your connection.',
        status: 0,
      });
    }

    return Promise.reject({
      message: error.message || 'An unexpected error occurred',
    });
  }
);

export default apiClient;
