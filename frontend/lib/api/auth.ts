import apiClient from './client';
import {
  RegisterRequest,
  LoginRequest,
  User,
  PasswordResetRequest,
  PasswordChangeRequest,
  EmailRequest,
  AddEmailRequest,
  TokenRequest,
} from '../types/auth';

export const authApi = {
  // Register new user
  register: async (data: RegisterRequest): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>('/auth/register', data);
    return response.data;
  },

  // Login user
  login: async (data: LoginRequest): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>('/auth/login', data);
    return response.data;
  },

  // Refresh session
  refresh: async (): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>('/auth/refresh');
    return response.data;
  },

  // Logout current session
  logout: async (): Promise<void> => {
    await apiClient.post('/auth/logout');
  },

  // Logout from all sessions
  logoutAll: async (): Promise<void> => {
    await apiClient.post('/auth/logout-all');
  },

  // Get current user
  getCurrentUser: async (): Promise<User> => {
    const response = await apiClient.get<User>('/auth/me');
    return response.data;
  },

  // Request password reset email
  requestPasswordReset: async (data: EmailRequest): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>('/auth/forgot-password', data);
    return response.data;
  },

  // Reset password with token
  resetPassword: async (data: PasswordResetRequest): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>('/auth/reset-password', data);
    return response.data;
  },

  // Change password (authenticated)
  changePassword: async (data: PasswordChangeRequest): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>('/auth/change-password', data);
    return response.data;
  },

  // Confirm email with token
  confirmEmail: async (data: TokenRequest): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>('/auth/confirm-email', data);
    return response.data;
  },

  // Resend confirmation email
  resendConfirmation: async (data: EmailRequest): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>('/auth/resend-confirmation', data);
    return response.data;
  },

  // Add/update email
  addEmail: async (data: AddEmailRequest): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>('/auth/me/email', data);
    return response.data;
  },
};
