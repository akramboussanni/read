package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/akramboussanni/gocode/internal/model"
	quiz "github.com/akramboussanni/gocode/internal/quiz"
	"github.com/akramboussanni/gocode/internal/utils"
	"github.com/go-chi/chi/v5"
)

func (ar *AdminRouter) HandleListCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := ar.Repos.Course.GetAllAdmin(r.Context())
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		return
	}
	if courses == nil {
		courses = []*model.Course{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

func (ar *AdminRouter) HandleCreateCourse(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Color       string `json:"color"`
		IsDefault   bool   `json:"is_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	now := time.Now().Unix()

	course := &model.Course{
		ID:          "course_" + strconv.FormatInt(now, 10),
		Title:       body.Title,
		Description: body.Description,
		Icon:        body.Icon,
		Color:       body.Color,
		IsDefault:   body.IsDefault,
		IsPublished: true,
		CreatedBy:   &user.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := ar.Repos.Course.Create(r.Context(), course); err != nil {
		http.Error(w, "Failed to create course", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(course)
}

func (ar *AdminRouter) HandleAutoGenerateCourse(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	var body struct {
		DeckID int64  `json:"deck_id,string"`
		Title  string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	course, err := ar.QuizService.AutoGenerateCourse(r.Context(), body.DeckID, body.Title, &user.ID)
	if err != nil {
		http.Error(w, "Failed to auto-generate course: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Load full course with nodes/edges
	nodes, _ := ar.Repos.CourseNode.GetByCourseID(r.Context(), course.ID)
	edges, _ := ar.Repos.CourseEdge.GetByCourseID(r.Context(), course.ID)
	course.Nodes = nodes
	course.Edges = edges

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(course)
}

func (ar *AdminRouter) HandleGetCourse(w http.ResponseWriter, r *http.Request) {
	courseID := chi.URLParam(r, "courseID")

	course, err := ar.Repos.Course.GetByID(r.Context(), courseID)
	if err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	nodes, _ := ar.Repos.CourseNode.GetByCourseID(r.Context(), courseID)
	edges, _ := ar.Repos.CourseEdge.GetByCourseID(r.Context(), courseID)
	course.Nodes = nodes
	course.Edges = edges

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(course)
}

func (ar *AdminRouter) HandleUpdateCourse(w http.ResponseWriter, r *http.Request) {
	courseID := chi.URLParam(r, "courseID")

	course, err := ar.Repos.Course.GetByID(r.Context(), courseID)
	if err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Color       string `json:"color"`
		IsDefault   bool   `json:"is_default"`
		IsPublished bool   `json:"is_published"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	course.Title = body.Title
	course.Description = body.Description
	course.Icon = body.Icon
	course.Color = body.Color
	course.IsDefault = body.IsDefault
	course.IsPublished = body.IsPublished

	if err := ar.Repos.Course.Update(r.Context(), course); err != nil {
		http.Error(w, "Failed to update course", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(course)
}

func (ar *AdminRouter) HandleDeleteCourse(w http.ResponseWriter, r *http.Request) {
	courseID := chi.URLParam(r, "courseID")

	if err := ar.Repos.Course.Delete(r.Context(), courseID); err != nil {
		http.Error(w, "Failed to delete course", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ar *AdminRouter) HandleAddNode(w http.ResponseWriter, r *http.Request) {
	courseID := chi.URLParam(r, "courseID")

	var body struct {
		NodeType      string  `json:"node_type"`
		Title         string  `json:"title"`
		Description   string  `json:"description"`
		Icon          string  `json:"icon"`
		PositionX     float64 `json:"position_x"`
		PositionY     float64 `json:"position_y"`
		SortOrder     int     `json:"sort_order"`
		QuizConfig    string  `json:"quiz_config"`
		LessonContent string  `json:"lesson_content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	now := importTime()

	var quizConfig *string
	if body.QuizConfig != "" {
		quizConfig = &body.QuizConfig
	}
	var lessonContent *string
	if body.LessonContent != "" {
		lessonContent = &body.LessonContent
	}

	node := &model.CourseNode{
		ID:            "node_" + strconv.FormatInt(now, 10),
		CourseID:      courseID,
		NodeType:      body.NodeType,
		Title:         body.Title,
		Description:   body.Description,
		Icon:          body.Icon,
		PositionX:     body.PositionX,
		PositionY:     body.PositionY,
		SortOrder:     body.SortOrder,
		QuizConfig:    quizConfig,
		LessonContent: lessonContent,
		CreatedAt:     now,
	}

	if err := ar.Repos.CourseNode.Create(r.Context(), node); err != nil {
		http.Error(w, "Failed to add node", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(node)
}

func (ar *AdminRouter) HandleUpdateNode(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "nodeID")

	node, err := ar.Repos.CourseNode.GetByID(r.Context(), nodeID)
	if err != nil {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	var body struct {
		Title         string  `json:"title"`
		Description   string  `json:"description"`
		Icon          string  `json:"icon"`
		PositionX     float64 `json:"position_x"`
		PositionY     float64 `json:"position_y"`
		SortOrder     int     `json:"sort_order"`
		QuizConfig    string  `json:"quiz_config"`
		LessonContent string  `json:"lesson_content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	node.Title = body.Title
	node.Description = body.Description
	node.Icon = body.Icon
	node.PositionX = body.PositionX
	node.PositionY = body.PositionY
	node.SortOrder = body.SortOrder
	if body.QuizConfig != "" {
		node.QuizConfig = &body.QuizConfig
	}
	if body.LessonContent != "" {
		node.LessonContent = &body.LessonContent
	}

	if err := ar.Repos.CourseNode.Update(r.Context(), node); err != nil {
		http.Error(w, "Failed to update node", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}

func (ar *AdminRouter) HandleDeleteNode(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "nodeID")

	if err := ar.Repos.CourseNode.Delete(r.Context(), nodeID); err != nil {
		http.Error(w, "Failed to delete node", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ar *AdminRouter) HandleAddEdge(w http.ResponseWriter, r *http.Request) {
	courseID := chi.URLParam(r, "courseID")

	var body struct {
		Source   string `json:"source"`
		Target   string `json:"target"`
		Label    string `json:"label"`
		EdgeType string `json:"edge_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.EdgeType == "" {
		body.EdgeType = "required"
	}

	now := importTime()
	edge := &model.CourseEdge{
		ID:           "edge_" + strconv.FormatInt(now, 10),
		CourseID:     courseID,
		SourceNodeID: body.Source,
		TargetNodeID: body.Target,
		Label:        body.Label,
		EdgeType:     body.EdgeType,
		CreatedAt:    now,
	}

	if err := ar.Repos.CourseEdge.Create(r.Context(), edge); err != nil {
		http.Error(w, "Failed to add edge", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(edge)
}

func (ar *AdminRouter) HandleDeleteEdge(w http.ResponseWriter, r *http.Request) {
	edgeID := chi.URLParam(r, "edgeID")

	if err := ar.Repos.CourseEdge.Delete(r.Context(), edgeID); err != nil {
		http.Error(w, "Failed to delete edge", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func importTime() int64 {
	return time.Now().Unix()
}

// HandleListTemplates returns available course templates
func (ar *AdminRouter) HandleListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := quiz.ListAvailableTemplates()
	if err != nil {
		http.Error(w, "Failed to list templates", http.StatusInternalServerError)
		return
	}
	if templates == nil {
		templates = []quiz.TemplateInfo{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

// HandleCreateFromTemplate creates a full course from a template file
func (ar *AdminRouter) HandleCreateFromTemplate(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	var body struct {
		TemplateFilename string `json:"template_filename"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.TemplateFilename == "" {
		http.Error(w, "template_filename is required", http.StatusBadRequest)
		return
	}

	course, err := ar.QuizService.CreateCourseFromTemplate(r.Context(), body.TemplateFilename, &user.ID)
	if err != nil {
		http.Error(w, "Failed to create course from template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Load full course with nodes/edges
	nodes, _ := ar.Repos.CourseNode.GetByCourseID(r.Context(), course.ID)
	edges, _ := ar.Repos.CourseEdge.GetByCourseID(r.Context(), course.ID)
	course.Nodes = nodes
	course.Edges = edges

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(course)
}
