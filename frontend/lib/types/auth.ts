// Auth request/response types matching backend API

export interface User {
  id: number;
  username: string;
  email: string;
  created_at: number;
  role: string;
  is_admin: boolean;
  email_confirmed: boolean;
}

export interface RegisterRequest {
  username: string;
  email?: string; // Optional - only needed for quiz creation
  password: string;
  url?: string; // Optional confirmation URL
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
}

export interface AddEmailRequest {
  email: string;
  url: string;
  password?: string; // Required only when changing a verified email
}

export interface EmailRequest {
  email: string;
  url?: string;
}

export interface TokenRequest {
  token: string;
  url?: string;
}

export interface PasswordResetRequest {
  token: string;
  new_password: string;
}

export interface PasswordChangeRequest {
  old_password: string;
  new_password: string;
}

export interface ApiError {
  error: string;
  message?: string;
}
