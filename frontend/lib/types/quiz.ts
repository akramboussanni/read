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
  id: string;
  deck_id: string;
  category_id: string;
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
  id: string;
  title: string;
  description: string;
  version?: number;
  deck_id: string;
  time_limit?: number;
  pass_percentage?: number;
  shuffle_questions?: boolean;
  question_mode?: string;
  difficulty?: string;
  is_public: boolean;
  created_by?: string;
  created_at?: number;
  updated_at?: number;
  category_selections?: CategorySelection[];
  gives_coins?: boolean;
  templates?: any[]; // To handle embedded templates
  coin_reward?: number;
  level_order?: number;
  is_dynamic?: boolean;
}

export interface CategorySelection {
  category_id: string;
  question_count: number;
}

export interface QuizAttempt {
  id: string;
  user_id: string;
  quiz_id: string;
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
  attempt_id: string;
  quiz_id: string;
  title: string;
  time_limit?: number;
  questions: QuestionWithOptions[];
  previous_answers?: SubmitAnswer[];
}

export interface QuestionWithOptions {
  id: string;
  question_text: string;
  question_type: string;
  direction?: string;
  points: number;
  options?: Option[];
}

export interface Option {
  id: string;
  option_text: string;
}

export interface SubmitQuizRequest {
  attempt_id: string;
  answers: SubmitAnswer[];
}

export interface SubmitAnswer {
  question_id: string;
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
  question_id: string;
  user_answer: string;
  correct_answer: string;
  is_correct: boolean;
  points_earned: number;
  ai_explanation?: string;
}

export interface CreateQuizRequest {
  title: string;
  description?: string;
  manual_questions?: ManualQuestionRequest[];
  auto_generate?: AutoGenerateRequest;
  is_public: boolean;
  // Admin-only fields
  pass_percentage?: number;
  gives_coins?: boolean;
  coin_reward?: number;
  level_order?: number;
  prerequisite_quiz_id?: string;
  is_system?: boolean;
  is_dynamic?: boolean;
}

export interface ManualQuestionRequest {
  question_text: string;
  correct_answer: string;
  options?: string[];
  question_type: 'mcq' | 'write_word' | 'translate';
  direction: 'source_to_target' | 'target_to_source';
}

export interface AutoGenerateRequest {
  deck_selections: DeckSelectionRequest[];
  question_types: string[];
  directions: string[];
  question_count: number;
  difficulty?: string;
}

export interface DeckSelectionRequest {
  deck_id: string;
  categories: string[];
}
