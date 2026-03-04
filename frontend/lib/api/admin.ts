import apiClient from './client';
import type {
  UserListResponse,
  UserDetailResponse,
  SystemStatsResponse,
  DeckWithCounts,
  CreateCourseRequest,
  AutoGenerateCourseRequest,
  UpdateCourseRequest,
  AddNodeRequest,
  AddEdgeRequest,
  TemplateInfo,
  CreateFromTemplateRequest,
} from '../types/admin';
import type { Course, CourseEdge, CourseNode } from '../types/course';

// User management
export async function getUsers(limit = 50, offset = 0): Promise<UserListResponse[]> {
  const response = await apiClient.get(`/admin/users?limit=${limit}&offset=${offset}`);
  return response.data;
}

export async function getUserDetail(userId: string): Promise<UserDetailResponse> {
  const response = await apiClient.get<UserDetailResponse>(`/admin/users/${userId}`);
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

// Course management
export async function listCoursesAdmin(): Promise<Course[]> {
  const response = await apiClient.get<Course[]>('/admin/courses');
  return response.data;
}

export async function getCourseAdmin(courseId: string): Promise<Course> {
  const response = await apiClient.get<Course>(`/admin/courses/${courseId}`);
  return response.data;
}

export async function createCourse(data: CreateCourseRequest): Promise<Course> {
  const response = await apiClient.post<Course>('/admin/courses', data);
  return response.data;
}

export async function autoGenerateCourse(data: AutoGenerateCourseRequest): Promise<Course> {
  const response = await apiClient.post<Course>('/admin/courses/auto-generate', data);
  return response.data;
}

export async function updateCourse(courseId: string, data: UpdateCourseRequest): Promise<Course> {
  const response = await apiClient.put<Course>(`/admin/courses/${courseId}`, data);
  return response.data;
}

export async function deleteCourse(courseId: string): Promise<void> {
  await apiClient.delete(`/admin/courses/${courseId}`);
}

// Course Nodes
export async function addCourseNode(courseId: string, data: AddNodeRequest): Promise<CourseNode> {
  const response = await apiClient.post<CourseNode>(`/admin/courses/${courseId}/nodes`, data);
  return response.data;
}

export async function updateCourseNode(nodeId: string, data: AddNodeRequest): Promise<CourseNode> {
  const response = await apiClient.put<CourseNode>(`/admin/courses/nodes/${nodeId}`, data);
  return response.data;
}

export async function deleteCourseNode(nodeId: string): Promise<void> {
  await apiClient.delete(`/admin/courses/nodes/${nodeId}`);
}

// Course Edges
export async function addCourseEdge(courseId: string, data: AddEdgeRequest): Promise<CourseEdge> {
  const response = await apiClient.post<CourseEdge>(`/admin/courses/${courseId}/edges`, data);
  return response.data;
}

export async function deleteCourseEdge(edgeId: string): Promise<void> {
  await apiClient.delete(`/admin/courses/edges/${edgeId}`);
}

export async function listDecks(): Promise<DeckWithCounts[]> {
  const response = await apiClient.get<DeckWithCounts[]>('/admin/decks');
  return response.data;
}

// Statistics
export async function getSystemStats(): Promise<SystemStatsResponse> {
  const response = await apiClient.get<SystemStatsResponse>('/admin/stats/overview');
  return response.data;
}

// Course Templates
export async function listTemplates(): Promise<TemplateInfo[]> {
  const response = await apiClient.get<TemplateInfo[]>('/admin/courses/templates');
  return response.data;
}

export async function createFromTemplate(data: CreateFromTemplateRequest): Promise<Course> {
  const response = await apiClient.post<Course>('/admin/courses/from-template', data);
  return response.data;
}
