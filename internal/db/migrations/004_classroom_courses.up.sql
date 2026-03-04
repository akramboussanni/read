CREATE TABLE classroom_courses (
    classroom_id BIGINT NOT NULL,
    course_id TEXT NOT NULL,
    added_at BIGINT NOT NULL,
    PRIMARY KEY (classroom_id, course_id),
    FOREIGN KEY (classroom_id) REFERENCES classrooms(id) ON DELETE CASCADE,
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE
);