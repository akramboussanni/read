import apiClient from './client';
import { Course, CourseStatus, UserEnrollment } from '../types/course';

export const courseApi = {
    listCourses: async (): Promise<Course[]> => {
        const response = await apiClient.get<Course[]>('/courses');
        return response.data;
    },

    getCourse: async (courseId: string): Promise<Course> => {
        const response = await apiClient.get<Course>(`/courses/${courseId}`);
        return response.data;
    },

    enroll: async (courseId: string): Promise<{ status: string }> => {
        const response = await apiClient.post<{ status: string }>(`/courses/${courseId}/enroll`);
        return response.data;
    },

    getMyEnrollments: async (): Promise<UserEnrollment[]> => {
        const response = await apiClient.get<UserEnrollment[]>('/courses/my/enrollments');
        return response.data;
    },

    setActiveCourse: async (courseId: string): Promise<{ status: string }> => {
        const response = await apiClient.put<{ status: string }>('/courses/my/active-course', { course_id: courseId });
        return response.data;
    },

    getCourseStatus: async (courseId: string): Promise<CourseStatus> => {
        const response = await apiClient.get<CourseStatus>(`/courses/${courseId}/status`);
        return response.data;
    },

    completeNode: async (courseId: string, nodeId: string): Promise<{ success: boolean }> => {
        const response = await apiClient.post<{ success: boolean }>('/quiz/nodes/complete', { course_id: courseId, node_id: nodeId });
        return response.data;
    },
};
