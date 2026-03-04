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
}

export interface SystemStatsResponse {
  total_users: number;
  active_users_7d: number;
}

// Decks
export interface DeckWithCounts {
  id: string;
  deck_key: string;
  title: string;
  description: string;
  deck_type: string;
  category_count: number;
  is_system: boolean;
}

export interface CreateCourseRequest {
  title: string;
  description: string;
  icon: string;
  color: string;
  is_default: boolean;
}

export interface AutoGenerateCourseRequest {
  deck_id: string;
  title: string;
}

export interface UpdateCourseRequest extends CreateCourseRequest {
  is_published: boolean;
}

export interface AddNodeRequest {
  node_type: string;
  title: string;
  description: string;
  icon: string;
  position_x: number;
  position_y: number;
  sort_order: number;
  quiz_config?: string;
  lesson_content?: string;
}

export interface AddEdgeRequest {
  source: string;
  target: string;
  label?: string;
  edge_type: string;
}

export interface TemplateInfo {
  filename: string;
  title: string;
  description: string;
  group_count: number;
  deck_key: string;
}

export interface CreateFromTemplateRequest {
  template_filename: string;
}

