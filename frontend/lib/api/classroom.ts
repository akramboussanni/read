import apiClient from './client';

export interface Classroom {
    id: string;
    name: string;
    description: string;
    teacher_id: string;
    join_code: string;
    is_locked: boolean;
    created_at: number;
}

export interface ClassroomAssignment {
    id: string;
    classroom_id: string;
    course_id: string;
    node_id: string;
    title: string;
    description: string;
    due_date: number;
    passing_grade: number;    // 0-100
    max_retakes: number;      // -1=unlimited, 0=no retakes, N=N retakes
    assignment_type: 'quiz' | 'path_progress'; // homework type
    created_at: number;
    is_completed?: boolean;
    score?: number;
    max_score?: number;
    score_percent?: number;
    completed_at?: number;
    course_name?: string;
    node_title?: string;
    completed_count?: number;
    total_students?: number;
}

export interface StudentStat {
    student_id: string;
    username: string;
    is_completed: boolean;
    score?: number;
    max_score?: number;
    percentage?: number;
    completed_at?: number;
    passed?: boolean;
}

export interface StudentAssignmentDetail {
    attempt_count: number;
    attempt: any | null;
    questions: any[];
    answers: any[];
    passed_assignment: boolean;
    passing_grade: number;
}

export const classroomApi = {
    listMyClasses: async () => {
        const response = await apiClient.get<{ teaching: Classroom[], enrolled: Classroom[] }>('/classroom/my');
        return response.data;
    },
    createClassroom: async (name: string, description: string) => {
        const response = await apiClient.post<Classroom>('/classroom/create', { name, description });
        return response.data;
    },
    joinClassroom: async (code: string) => {
        const response = await apiClient.post<{ status: string, name: string, id: string }>(`/classroom/join/${code}`);
        return response.data;
    },
    getClassDetails: async (id: string) => {
        const response = await apiClient.get<any>(`/classroom/${id}`);
        return response.data;
    },
    updateClass: async (id: string, data: Partial<Classroom>) => {
        const response = await apiClient.put<Classroom>(`/classroom/${id}`, data);
        return response.data;
    },
    createAssignment: async (classId: string, data: {
        course_id: string;
        node_id: string;
        title: string;
        description?: string;
        due_date?: number;
        passing_grade?: number;
        max_retakes?: number;
        assignment_type?: 'quiz' | 'path_progress';
    }) => {
        const response = await apiClient.post<any>(`/classroom/${classId}/assignments`, data);
        return response.data;
    },
    updateAssignment: async (classId: string, asgnId: string, data: {
        title: string;
        description?: string;
        due_date?: number;
        passing_grade: number;
        max_retakes: number;
    }) => {
        const response = await apiClient.put<any>(`/classroom/${classId}/assignments/${asgnId}`, data);
        return response.data;
    },
    deleteAssignment: async (classId: string, asgnId: string) => {
        await apiClient.delete(`/classroom/${classId}/assignments/${asgnId}`);
    },
    removeStudent: async (classId: string, studentId: string) => {
        await apiClient.delete(`/classroom/${classId}/students/${studentId}`);
    },
    getAssignmentStats: async (classId: string, asgnId: string) => {
        const response = await apiClient.get<StudentStat[]>(`/classroom/${classId}/assignments/${asgnId}/stats`);
        return response.data;
    },
    getStudentAssignmentDetail: async (classId: string, asgnId: string, studentId: string): Promise<StudentAssignmentDetail> => {
        const response = await apiClient.get<StudentAssignmentDetail>(
            `/classroom/${classId}/assignments/${asgnId}/students/${studentId}`
        );
        return response.data;
    },
    addCourse: async (classId: string, courseId: string) => {
        await apiClient.post(`/classroom/${classId}/courses`, { course_id: courseId });
    },
    removeCourse: async (classId: string, courseId: string) => {
        await apiClient.delete(`/classroom/${classId}/courses/${courseId}`);
    }
};
