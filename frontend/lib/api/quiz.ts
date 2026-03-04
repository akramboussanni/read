import apiClient from './client';
import {
  Quiz,
  QuizListResponse,
  StartQuizResponse,
  Deck,
  Category,
  CreateQuizRequest,
  AutoGenerateRequest,
  ManualQuestionRequest,
} from '../types/quiz';

export const quizApi = {
  // List all quizzes (public)
  listQuizzes: async (params?: {
    page?: number;
    page_size?: number;
  }): Promise<QuizListResponse> => {
    const response = await apiClient.get<QuizListResponse>('/quiz/', { params });
    return response.data;
  },

  // Get quiz by ID
  getQuiz: async (quizId: string): Promise<Quiz & { templates?: any[] }> => {
    const response = await apiClient.get<any>(`/quiz/${quizId}`);
    if (response.data && response.data.quiz) {
      return { ...response.data.quiz, templates: response.data.templates };
    }
    return response.data;
  },

  // Get user's quizzes (authenticated)
  getMyQuizzes: async (): Promise<Quiz[]> => {
    const response = await apiClient.get<Quiz[]>('/quiz/my');
    return response.data || [];
  },

  // Create new quiz (authenticated)
  createQuiz: async (data: CreateQuizRequest): Promise<Quiz> => {
    const response = await apiClient.post<Quiz>('/quiz/', data);
    return response.data;
  },

  // Update quiz (authenticated)
  updateQuiz: async (quizId: string, data: any): Promise<Quiz> => {
    const response = await apiClient.put<any>(`/quiz/${quizId}`, data);
    if (response.data && response.data.quiz) {
      return { ...response.data.quiz, templates: response.data.templates };
    }
    return response.data;
  },

  // Delete quiz (authenticated)
  deleteQuiz: async (quizId: string): Promise<void> => {
    await apiClient.delete(`/quiz/${quizId}`);
  },

  // Start quiz attempt (authenticated)
  startQuiz: async (
    quizId: string,
    courseId?: string,
    nodeId?: string,
    assignmentId?: string
  ): Promise<StartQuizResponse> => {
    const response = await apiClient.post<StartQuizResponse>('/quiz/start', {
      quiz_id: quizId,
      course_id: courseId,
      node_id: nodeId,
      assignment_id: assignmentId,
    });
    return response.data;
  },

  // Submit single answer (authenticated)
  submitAnswer: async (data: {
    attempt_id: string;
    question_id: string;
    answer: string;
  }): Promise<{
    is_correct: boolean;
    correct_answer: string;
    points_earned: number;
    needs_more_detail: boolean;
    ai_explanation?: string;
  }> => {
    const response = await apiClient.post('/quiz/answer', data);
    return response.data;
  },

  // Complete quiz attempt (authenticated)
  completeQuiz: async (attemptId: string): Promise<{
    attempt_id: string;
    score: number;
    max_score: number;
    percentage: number;
    passed: boolean;
    coins_earned: number;
    time_taken: number;
    results: Array<{
      question_id: string;
      question_text: string;
      correct_answer: string;
      user_answer: string;
      is_correct: boolean;
      points_earned: number;
      ai_explanation?: string;
    }>;
  }> => {
    const response = await apiClient.post('/quiz/complete', {
      attempt_id: attemptId,
    });
    return response.data;
  },

  // Get attempt details
  getAttempt: async (attemptId: string): Promise<any> => {
    const response = await apiClient.get(`/quiz/attempt/${attemptId}`);
    return response.data;
  },

  // Get quiz history
  getHistory: async (limit?: number): Promise<any[]> => {
    const response = await apiClient.get('/quiz/my/history', {
      params: { limit },
    });
    return response.data;
  },

  // List all decks (authenticated)
  listDecks: async (): Promise<Deck[]> => {
    const response = await apiClient.get<Deck[]>('/quiz/decks');
    return response.data;
  },

  // Get categories for a deck (authenticated)
  getCategories: async (deckId: string): Promise<Category[]> => {
    const response = await apiClient.get<Category[]>(
      `/quiz/decks/${deckId}/categories`
    );
    return response.data;
  },

  // Generate preview questions for manual selection
  generatePreview: async (
    data: AutoGenerateRequest
  ): Promise<ManualQuestionRequest[]> => {
    const response = await apiClient.post<ManualQuestionRequest[]>(
      '/quiz/preview',
      data
    );
    return response.data;
  },
};
