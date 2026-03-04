export interface Course {
    id: string;
    title: string;
    description: string;
    icon: string;
    color: string;
    deck_id?: string;
    is_default: boolean;
    is_published: boolean;
    created_by?: string;
    created_at: string;
    updated_at: string;
    nodes?: CourseNode[];
    edges?: CourseEdge[];
}

export interface CourseNode {
    id: string;
    course_id: string;
    node_type: 'lesson' | 'quiz' | 'milestone' | 'start' | 'checkpoint' | 'reference';
    title: string;
    description: string;
    icon: string;
    position_x: number;
    position_y: number;
    sort_order: number;
    quiz_config?: any;
    lesson_content?: any;
    metadata?: any;
    created_at: string;
}

export interface CourseEdge {
    id: string;
    course_id: string;
    source: string;
    target: string;
    label?: string;
    edge_type: 'required' | 'optional' | 'bonus';
    created_at: string;
}

export interface UserEnrollment {
    id: string;
    user_id: string;
    course_id: string;
    status: 'active' | 'completed' | 'paused';
    progress: number;
    current_node_id?: string;
    completed_nodes: string[];
    enrolled_at: string;
    completed_at?: string;
    last_accessed_at: string;
    course?: Course;
}

export interface CourseNodeStatus {
    node_id: string;
    node_type: string;
    state: 'locked' | 'unlocked' | 'mastered' | 'completed';
}

export interface CourseStatus {
    node_statuses: Record<string, CourseNodeStatus>;
    enrollment?: UserEnrollment;
}
