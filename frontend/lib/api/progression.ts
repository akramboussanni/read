import apiClient from './client';
import {
  ProgressionStatus,
  QuizProgress,
  QuizPreview,
} from '../types/progression';
import { StartQuizResponse, SubmitQuizRequest, SubmitQuizResponse } from '../types/quiz';

export const progressionApi = {
  // Get user progression status
  getStatus: async (): Promise<ProgressionStatus> => {
    const response = await apiClient.get<ProgressionStatus>('/progression/status');
    return response.data;
  },

  // List progression quizzes
  listQuizzes: async (): Promise<QuizPreview[]> => {
    const response = await apiClient.get<QuizPreview[]>('/progression/quizzes');
    return response.data;
  },

  // Get specific quiz progress
  getQuizProgress: async (quizId: number): Promise<QuizProgress> => {
    const response = await apiClient.get<QuizProgress>(`/progression/quiz/${quizId}`);
    return response.data;
  },

  // Start progression quiz
  startQuiz: async (quizId: number): Promise<StartQuizResponse> => {
    const response = await apiClient.post<StartQuizResponse>(
      `/progression/quiz/${quizId}/start`
    );
    return response.data;
  },

  // Submit progression quiz
  submitQuiz: async (
    quizId: number,
    data: SubmitQuizRequest
  ): Promise<SubmitQuizResponse> => {
    const response = await apiClient.post<SubmitQuizResponse>(
      `/progression/quiz/${quizId}/submit`,
      data
    );
    return response.data;
  },
};
