// Admin-specific types matching backend API

export interface UserListResponse {
  id: string;
  username: string;
  email: string;
  role: string;
  is_admin: boolean;
  email_confirmed: boolean;
  created_at: string;
}

export interface UserDetailResponse {
  user: UserListResponse;
  current_level: number;
  total_quizzes_completed: number;
  total_coins_earned: number;
  coin_balance: number;
  streak_days: number;
  last_activity_date?: string;
}

export interface QuizStatsResponse {
  quiz_id: string;
  title: string;
  version?: number;
  created_by?: string;
  creator_username?: string;
  is_system: boolean;
  total_attempts: number;
  average_score: number;
  pass_rate: number;
  created_at: string;
}

export interface SystemStatsResponse {
  total_users: number;
  total_quizzes: number;
  total_attempts: number;
  active_users_7d: number;
  average_score?: number;
  completion_rate?: number;
  total_questions?: number;
}

export interface UpdateQuizRequest {
  title?: string;
  description?: string;
  pass_percentage?: number;
  shuffle_questions?: boolean;
  gives_coins?: boolean;
  coin_reward?: number;
  level_order?: number;
  prerequisite_quiz_id?: string;
  is_public?: boolean;
}

export interface CreateQuizRequest {
  title: string;
  description?: string;
  manual_questions?: ManualQuestion[];
  auto_generate?: AutoGenerateConfig;
  is_public: boolean;
  // Admin-only fields
  pass_percentage?: number;
  gives_coins?: boolean;
  coin_reward?: number;
  level_order?: number;
  prerequisite_quiz_id?: string;
  is_system?: boolean;
}

export interface ManualQuestion {
  question_text: string;
  correct_answer: string;
  options?: string[];
  question_type: 'mcq' | 'write_word' | 'translate';
  direction: 'source_to_target' | 'target_to_source';
}

export interface AutoGenerateConfig {
  deck_selections: DeckSelection[];
  question_types: ('mcq' | 'write_word' | 'translate')[];
  directions: ('source_to_target' | 'target_to_source')[];
  question_count: number;
  difficulty?: string;
}

export interface DeckSelection {
  deck_id: number;
  categories: string[];
}

export interface UserQuizResponse {
  id: string;
  title: string;
  description: string;
  created_by: string;
  creator_username: string;
  is_public: boolean;
  total_attempts: number;
  created_at: string;
}
