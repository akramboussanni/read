// Quiz types matching backend API

export interface Deck {
  id: string;
  deck_key: string;
  title: string;
  version?: number;
  source_file?: string;
  is_system?: boolean;
  created_at?: number;
  category_count?: number;
  question_count?: number;
}

export interface Category {
  id: string;
  deck_id: string;
  category_key: string;
  title: string;
  display_order?: number;
  created_at?: number;
  question_count?: number;
}

export interface Question {
  id: number;
  deck_id: number;
  category_id: number;
  question_key: string;
  question_text: string;
  correct_answer: string;
  arabic: string;
  french: string;
  question_type: string;
  difficulty?: string;
  points: number;
  hint?: string;
  explanation?: string;
  created_at: number;
  updated_at?: number;
  is_active: boolean;
}

export interface Quiz {
  id: number;
  title: string;
  description: string;
  version?: number;
  deck_id: number;
  time_limit?: number;
  pass_percentage?: number;
  shuffle_questions?: boolean;
  question_mode?: string;
  difficulty?: string;
  is_public: boolean;
  created_by?: number;
  created_at?: number;
  updated_at?: number;
  category_selections?: CategorySelection[];
}

export interface CategorySelection {
  category_id: number;
  question_count: number;
}

export interface QuizAttempt {
  id: number;
  user_id: number;
  quiz_id: number;
  score: number;
  max_score: number;
  started_at: number;
  completed_at?: number;
  time_taken?: number;
}

export interface QuizListResponse {
  quizzes: Quiz[];
  total: number;
  page: number;
  page_size: number;
}

export interface StartQuizResponse {
  attempt_id: number;
  quiz_id: number;
  title: string;
  time_limit?: number;
  questions: QuestionWithOptions[];
}

export interface QuestionWithOptions {
  id: number;
  question_text: string;
  question_type: string;
  points: number;
  options?: Option[];
}

export interface Option {
  id: number;
  option_text: string;
}

export interface SubmitQuizRequest {
  attempt_id: number;
  answers: SubmitAnswer[];
}

export interface SubmitAnswer {
  question_id: number;
  answer: string;
}

export interface SubmitQuizResponse {
  score: number;
  max_score: number;
  percentage: number;
  passed: boolean;
  coins_earned: number;
  next_unlocked?: number;
  results: AnswerResult[];
}

export interface AnswerResult {
  question_id: number;
  user_answer: string;
  correct_answer: string;
  is_correct: boolean;
  points_earned: number;
}
