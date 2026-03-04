package model

// ProgressionPath is now called Course
type Course struct {
	ID          string `db:"id" json:"id"`
	Title       string `db:"title" json:"title"`
	Description string `db:"description" json:"description"`
	Icon        string `db:"icon" json:"icon"`
	Color       string `db:"color" json:"color"`
	DeckID      *int64 `db:"deck_id" json:"deck_id,omitempty"`
	IsDefault   bool   `db:"is_default" json:"is_default"`
	IsPublished bool   `db:"is_published" json:"is_published"`
	CreatedBy   *int64 `db:"created_by" json:"created_by,string,omitempty"`
	CreatedAt   int64  `db:"created_at" json:"created_at"`
	UpdatedAt   int64  `db:"updated_at" json:"updated_at"`
	// Populated in API responses
	Nodes []*CourseNode `json:"nodes,omitempty"`
	Edges []*CourseEdge `json:"edges,omitempty"`
}

// CourseNode represents a node in the course tree
type CourseNode struct {
	ID            string  `db:"id" json:"id"`
	CourseID      string  `db:"course_id" json:"course_id"`
	NodeType      string  `db:"node_type" json:"node_type"` // 'lesson','quiz','milestone','start','checkpoint'
	Title         string  `db:"title" json:"title"`
	Description   string  `db:"description" json:"description"`
	Icon          string  `db:"icon" json:"icon"`
	PositionX     float64 `db:"position_x" json:"position_x"`
	PositionY     float64 `db:"position_y" json:"position_y"`
	SortOrder     int     `db:"sort_order" json:"sort_order"`
	QuizConfig    *string `db:"quiz_config" json:"quiz_config,omitempty"`
	LessonContent *string `db:"lesson_content" json:"lesson_content,omitempty"`
	Metadata      *string `db:"metadata" json:"metadata,omitempty"`
	CreatedAt     int64   `db:"created_at" json:"created_at"`
}

// CourseEdge represents a connection between course nodes
type CourseEdge struct {
	ID           string `db:"id" json:"id"`
	CourseID     string `db:"course_id" json:"course_id"`
	SourceNodeID string `db:"source_node_id" json:"source"`
	TargetNodeID string `db:"target_node_id" json:"target"`
	Label        string `db:"label" json:"label"`
	EdgeType     string `db:"edge_type" json:"edge_type"` // 'required','optional','bonus'
	CreatedAt    int64  `db:"created_at" json:"created_at"`
}

// UserEnrollment tracks user progress in a course
type UserEnrollment struct {
	ID             string  `db:"id" json:"id"`
	UserID         int64   `db:"user_id" json:"user_id,string"`
	CourseID       string  `db:"course_id" json:"course_id"`
	Status         string  `db:"status" json:"status"` // 'active','completed','paused'
	Progress       float64 `db:"progress" json:"progress"`
	CurrentNodeID  *string `db:"current_node_id" json:"current_node_id,omitempty"`
	CompletedNodes string  `db:"completed_nodes" json:"completed_nodes"` // JSON array
	EnrolledAt     int64   `db:"enrolled_at" json:"enrolled_at"`
	CompletedAt    *int64  `db:"completed_at" json:"completed_at,omitempty"`
	LastAccessedAt int64   `db:"last_accessed_at" json:"last_accessed_at"`
}
