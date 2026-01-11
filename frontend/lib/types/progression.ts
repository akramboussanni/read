// Progression tracking types

export interface ProgressionStatus {
  current_level: number;
  total_quizzes_completed: number;
  total_coins_earned: number;
  streak_days: number;
  coin_balance: number;
  next_quiz?: QuizPreview;
}

export interface QuizProgress {
  quiz_id: number;
  quiz_title: string;
  attempts: number;
  best_score: number;
  average_score: number;
  last_attempt?: number;
  completed: boolean;
}

export interface QuizPreview {
  id: number;
  title: string;
  description: string;
  level: number;
  is_locked: boolean;
  is_completed: boolean;
  best_score?: number;
  best_percentage?: number;
  coin_reward: number;
  question_count: number;
}
