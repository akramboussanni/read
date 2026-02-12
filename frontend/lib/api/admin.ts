import apiClient from './client';
import type {
  UserListResponse,
  UserDetailResponse,
  QuizStatsResponse,
  SystemStatsResponse,
  UpdateQuizRequest,
  CreateQuizRequest,
  UserQuizResponse,
} from '../types/admin';

// User management
export async function getUsers(limit = 50, offset = 0): Promise<UserListResponse[]> {
  const response = await apiClient.get(`/admin/users?limit=${limit}&offset=${offset}`);
  return response.data;
}

export async function getUserDetail(userId: string): Promise<UserDetailResponse> {
  const response = await apiClient.get(`/admin/users/${userId}`);
  return response.data;
}

export async function changeUserPassword(userId: string, newPassword: string): Promise<void> {
  await apiClient.post(`/admin/users/${userId}/password`, {
    user_id: userId,
    new_password: newPassword,
  });
}

export async function deleteUser(userId: string): Promise<void> {
  await apiClient.delete(`/admin/users/${userId}`);
}

// Quiz management
export async function createQuiz(data: CreateQuizRequest): Promise<{ quiz_id: string }> {
  const response = await apiClient.post('/admin/quizzes', data);
  return response.data;
}

export async function updateQuiz(quizId: string, data: UpdateQuizRequest): Promise<void> {
  await apiClient.put(`/admin/quizzes/${quizId}`, data);
}

export async function deleteQuiz(quizId: string): Promise<void> {
  await apiClient.delete(`/admin/quizzes/${quizId}`);
}

export async function getQuizStats(limit = 50, offset = 0): Promise<QuizStatsResponse[]> {
  const response = await apiClient.get(`/admin/quizzes/stats?limit=${limit}&offset=${offset}`);
  return response.data;
}

export async function getUserGeneratedQuizzes(limit = 50, offset = 0): Promise<UserQuizResponse[]> {
  const response = await apiClient.get(`/admin/quizzes/user-generated?limit=${limit}&offset=${offset}`);
  return response.data;
}

// Statistics
export async function getSystemStats(): Promise<SystemStatsResponse> {
  const response = await apiClient.get('/admin/stats/overview');
  return response.data;
}
