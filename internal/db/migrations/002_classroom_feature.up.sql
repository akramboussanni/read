-- Migration for Classroom feature
CREATE TABLE classrooms (
    id BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    teacher_id BIGINT NOT NULL,
    join_code VARCHAR(12) UNIQUE NOT NULL,
    is_locked BOOLEAN NOT NULL DEFAULT FALSE, -- Teachers can "lock" enrollment
    created_at BIGINT NOT NULL,
    FOREIGN KEY (teacher_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE classroom_students (
    classroom_id BIGINT NOT NULL,
    student_id BIGINT NOT NULL,
    joined_at BIGINT NOT NULL,
    PRIMARY KEY (classroom_id, student_id),
    FOREIGN KEY (classroom_id) REFERENCES classrooms(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE classroom_assignments (
    id BIGINT PRIMARY KEY,
    classroom_id BIGINT NOT NULL,
    course_id TEXT NOT NULL,
    node_id TEXT NOT NULL, -- Pointing to a specific lesson or quiz
    title TEXT NOT NULL,
    description TEXT,
    due_date BIGINT,
    created_at BIGINT NOT NULL,
    FOREIGN KEY (classroom_id) REFERENCES classrooms(id) ON DELETE CASCADE,
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
    FOREIGN KEY (node_id) REFERENCES course_nodes(id) ON DELETE CASCADE
);

CREATE INDEX idx_classrooms_teacher ON classrooms(teacher_id);
CREATE INDEX idx_classroom_students_student ON classroom_students(student_id);
CREATE INDEX idx_classroom_assignments_class ON classroom_assignments(classroom_id);
