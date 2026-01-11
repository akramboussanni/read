import apiClient from './client';
import {
  Quiz,
  QuizListResponse,
  StartQuizResponse,
  SubmitQuizRequest,
  SubmitQuizResponse,
  Deck,
  Category,
} from '../types/quiz';

export const quizApi = {
  // List all quizzes (public)
  listQuizzes: async (params?: {
    page?: number;
    page_size?: number;
    difficulty?: string;
  }): Promise<QuizListResponse> => {
    const response = await apiClient.get<QuizListResponse>('/quiz/', { params });
    return response.data;
  },

  // Get quiz by ID (public)
  getQuiz: async (quizId: number): Promise<Quiz> => {
    const response = await apiClient.get<Quiz>(`/quiz/${quizId}`);
    return response.data;
  },

  // Get user's quizzes (authenticated)
  getMyQuizzes: async (): Promise<Quiz[]> => {
    const response = await apiClient.get<Quiz[]>('/quiz/my');
    return response.data;
  },

  // Create new quiz (authenticated)
  createQuiz: async (quiz: Partial<Quiz>): Promise<Quiz> => {
    const response = await apiClient.post<Quiz>('/quiz/', quiz);
    return response.data;
  },

  // Update quiz (authenticated)
  updateQuiz: async (quizId: number, quiz: Partial<Quiz>): Promise<Quiz> => {
    const response = await apiClient.put<Quiz>(`/quiz/${quizId}`, quiz);
    return response.data;
  },

  // Delete quiz (authenticated)
  deleteQuiz: async (quizId: number): Promise<void> => {
    await apiClient.delete(`/quiz/${quizId}`);
  },

  // Start quiz attempt (authenticated)
  startQuiz: async (quizId: number): Promise<StartQuizResponse> => {
    const response = await apiClient.post<StartQuizResponse>(`/quiz/${quizId}/start`);
    return response.data;
  },

  // Submit quiz answers (authenticated)
  submitQuiz: async (
    quizId: number,
    data: SubmitQuizRequest
  ): Promise<SubmitQuizResponse> => {
    const response = await apiClient.post<SubmitQuizResponse>(
      `/quiz/${quizId}/submit`,
      data
    );
    return response.data;
  },

  // List all decks (authenticated)
  listDecks: async (): Promise<Deck[]> => {
    const response = await apiClient.get<Deck[]>('/quiz/decks');
    return response.data;
  },

  // Get categories for a deck (authenticated)
  getCategories: async (deckId: string): Promise<Category[]> => {
    const response = await apiClient.get<Category[]>(`/quiz/decks/${deckId}/categories`);
    return response.data;
  },
};
