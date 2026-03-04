package quiz

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"time"

	"github.com/akramboussanni/gocode/internal/applog"
	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/utils"
)

//go:embed data/templates/*.json
var embeddedTemplates embed.FS

// ============================================================
// COURSE TEMPLATE JSON STRUCTURES
// ============================================================

// CourseTemplate represents a complete course definition from JSON
type CourseTemplate struct {
	TemplateVersion int                   `json:"template_version"`
	Course          CourseTemplateMeta    `json:"course"`
	Groups          []CourseTemplateGroup `json:"groups"`
}

// CourseTemplateMeta holds the course-level metadata
type CourseTemplateMeta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	IsDefault   bool   `json:"is_default"`
	DeckKey     string `json:"deck_key"`
}

// CourseTemplateGroup represents a thematic group in the course
type CourseTemplateGroup struct {
	Name        string                   `json:"name"`
	Categories  []string                 `json:"categories"`
	Reference   *CourseTemplateReference `json:"reference,omitempty"`
	QuizReward  int                      `json:"quiz_reward,omitempty"`  // Reward for each individual category quiz
	SynthReward int                      `json:"synth_reward,omitempty"` // Reward for the synthesis quiz (if multi-category)
}

// CourseTemplateReference is the rich lesson content shown alongside quizzes
type CourseTemplateReference struct {
	Title   string          `json:"title"`
	Content json.RawMessage `json:"content"` // JSON lesson blocks (text, flashcards, quran, quiz, tip)
}

// ============================================================
// LISTING AVAILABLE TEMPLATES
// ============================================================

// TemplateInfo is a summary returned to the frontend
type TemplateInfo struct {
	Filename    string `json:"filename"`
	Title       string `json:"title"`
	Description string `json:"description"`
	GroupCount  int    `json:"group_count"`
	DeckKey     string `json:"deck_key"`
}

// ListAvailableTemplates returns all embedded course templates
func ListAvailableTemplates() ([]TemplateInfo, error) {
	fsys, err := fs.Sub(embeddedTemplates, "data/templates")
	if err != nil {
		return nil, fmt.Errorf("failed to access templates: %w", err)
	}

	var templates []TemplateInfo

	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil
		}

		var tmpl CourseTemplate
		if err := json.Unmarshal(data, &tmpl); err != nil {
			return nil
		}

		templates = append(templates, TemplateInfo{
			Filename:    path,
			Title:       tmpl.Course.Title,
			Description: tmpl.Course.Description,
			GroupCount:  len(tmpl.Groups),
			DeckKey:     tmpl.Course.DeckKey,
		})

		return nil
	})

	return templates, err
}

// ============================================================
// CREATE COURSE FROM TEMPLATE
// ============================================================

// CreateCourseFromTemplate builds a complete course from a template file
func (s *QuizService) CreateCourseFromTemplate(ctx context.Context, templateFilename string, createdBy *int64) (*model.Course, error) {
	// 1. Load the template
	fsys, err := fs.Sub(embeddedTemplates, "data/templates")
	if err != nil {
		return nil, fmt.Errorf("failed to access templates: %w", err)
	}

	data, err := fs.ReadFile(fsys, templateFilename)
	if err != nil {
		return nil, fmt.Errorf("template file not found: %s", templateFilename)
	}

	var tmpl CourseTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("invalid template JSON: %w", err)
	}

	return s.CreateCourseFromTemplateData(ctx, &tmpl, createdBy)
}

// CreateCourseFromTemplateData creates a course from parsed template data
func (s *QuizService) CreateCourseFromTemplateData(ctx context.Context, tmpl *CourseTemplate, createdBy *int64) (*model.Course, error) {
	// 2. Find the deck by key
	deck, err := s.repos.Deck.GetByKey(ctx, tmpl.Course.DeckKey)
	if err != nil {
		return nil, fmt.Errorf("deck not found for key %q — make sure the vocabulary has been seeded: %w", tmpl.Course.DeckKey, err)
	}

	// 3. Load categories from DB, build key → ID map
	dbCategories, err := s.repos.Category.GetByDeckID(ctx, deck.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load categories: %w", err)
	}
	categoryMap := make(map[string]int64)
	for _, cat := range dbCategories {
		categoryMap[cat.CategoryKey] = cat.ID
	}

	now := time.Now().Unix()
	courseID := fmt.Sprintf("course_tmpl_%d", now)

	// 4. Create the course
	course := &model.Course{
		ID:          courseID,
		Title:       tmpl.Course.Title,
		Description: tmpl.Course.Description,
		Icon:        tmpl.Course.Icon,
		Color:       tmpl.Course.Color,
		DeckID:      &deck.ID,
		IsDefault:   tmpl.Course.IsDefault,
		IsPublished: true,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repos.Course.Create(ctx, course); err != nil {
		return nil, fmt.Errorf("failed to create course: %w", err)
	}

	// 5. Create start node
	startNode := &model.CourseNode{
		ID:        courseID + "_start",
		CourseID:  courseID,
		NodeType:  "start",
		Title:     "Départ",
		PositionX: 250,
		PositionY: 0,
		SortOrder: 0,
		CreatedAt: now,
	}
	if err := s.repos.CourseNode.Create(ctx, startNode); err != nil {
		return nil, fmt.Errorf("failed to create start node: %w", err)
	}

	prevNodeIDs := []string{startNode.ID}
	currentYPos := 0.0
	edgeCounter := 0

	for gIdx, group := range tmpl.Groups {
		// A. Check all categories exist in DB
		var validCatIDs []int64
		for _, catKey := range group.Categories {
			if id, ok := categoryMap[catKey]; ok {
				validCatIDs = append(validCatIDs, id)
			} else {
				applog.Warn("Template category %q not found in deck — skipping", catKey)
			}
		}
		if len(validCatIDs) == 0 {
			continue
		}

		// B. Create milestone for multi-category groups
		var milestoneID string
		if len(group.Categories) > 1 {
			currentYPos += 140.0
			milestoneID = fmt.Sprintf("%s_mile_%d", courseID, gIdx)
			milestone := &model.CourseNode{
				ID:          milestoneID,
				CourseID:    courseID,
				NodeType:    "milestone",
				Title:       group.Name,
				Description: fmt.Sprintf("Chapitre : %s", group.Name),
				Icon:        "folder",
				PositionX:   250,
				PositionY:   currentYPos,
				SortOrder:   gIdx * 10,
				CreatedAt:   now,
			}
			_ = s.repos.CourseNode.Create(ctx, milestone)

			for _, pID := range prevNodeIDs {
				edgeCounter++
				edge := &model.CourseEdge{
					ID:           fmt.Sprintf("%s_e_%d", courseID, edgeCounter),
					CourseID:     courseID,
					SourceNodeID: pID,
					TargetNodeID: milestoneID,
					EdgeType:     "required",
					CreatedAt:    now,
				}
				_ = s.repos.CourseEdge.Create(ctx, edge)
			}
		}

		// C. Create REFERENCE node as an optional sibling at the SAME level as quizzes
		// User can access it before, during, or after the quiz — it never blocks progress
		currentYPos += 140.0
		var refSiblingID string
		if group.Reference != nil && len(group.Reference.Content) > 0 {
			refSiblingID = fmt.Sprintf("%s_ref_%d", courseID, gIdx)

			// Wrap content in a lesson envelope
			envelope, _ := json.Marshal(map[string]interface{}{
				"type":   "lesson",
				"title":  group.Reference.Title,
				"blocks": group.Reference.Content, // already valid JSON
			})
			contentStr := string(envelope)

			// Position to the far left, quizzes spread to the right
			refNode := &model.CourseNode{
				ID:            refSiblingID,
				CourseID:      courseID,
				NodeType:      "reference",
				Title:         group.Name,
				Description:   group.Reference.Title,
				Icon:          "book-open",
				PositionX:     -150, // far left — quizzes are centred at 250
				PositionY:     currentYPos,
				SortOrder:     gIdx*10 + 1,
				LessonContent: &contentStr,
				CreatedAt:     now,
			}
			_ = s.repos.CourseNode.Create(ctx, refNode)

			// Optional edge from milestone or previous nodes
			connectFrom := prevNodeIDs
			if milestoneID != "" {
				connectFrom = []string{milestoneID}
			}
			for _, pID := range connectFrom {
				edgeCounter++
				e := &model.CourseEdge{
					ID:           fmt.Sprintf("%s_e_%d", courseID, edgeCounter),
					CourseID:     courseID,
					SourceNodeID: pID,
					TargetNodeID: refSiblingID,
					EdgeType:     "optional", // dashed line, no lock
					CreatedAt:    now,
				}
				_ = s.repos.CourseEdge.Create(ctx, e)
			}
		}

		// D. Create quiz nodes for each category (same Y level as reference)
		var currentGroupEndNodes []string

		for cIdx, catKey := range group.Categories {
			catID, ok := categoryMap[catKey]
			if !ok {
				continue
			}

			// Find category title from DB
			catTitle := catKey
			for _, dbCat := range dbCategories {
				if dbCat.CategoryKey == catKey {
					catTitle = dbCat.Title
					break
				}
			}

			// X position — shift right if reference sibling exists on the left
			xBase := 250.0
			if refSiblingID != "" {
				xBase = 350.0 // shift quiz cluster right to make room
			}
			offsetX := float64(cIdx) - float64(len(group.Categories)-1)/2.0
			posX := xBase + (offsetX * 260.0)

			actualQuizID := utils.GenerateSnowflakeID()
			quizNodeID := fmt.Sprintf("%s_q_%d_%d", courseID, gIdx, cIdx)

			quizNodeConfig := map[string]interface{}{
				"deck_id":        deck.ID,
				"category_id":    catID,
				"quiz_id":        strconv.FormatInt(actualQuizID, 10),
				"question_types": []string{"mcq", "translate"},
				"directions":     []string{"source_to_target", "target_to_source", "attach_suffix", "conjugate"},
				"question_count": 10,
				"coin_reward":    utils.IfElse(group.QuizReward > 0, group.QuizReward, 10),
			}
			configJSON, _ := json.Marshal(quizNodeConfig)
			configStr := string(configJSON)

			quizNode := &model.CourseNode{
				ID:          quizNodeID,
				CourseID:    courseID,
				NodeType:    "quiz",
				Title:       catTitle,
				Description: fmt.Sprintf("Apprends %s", catTitle),
				Icon:        "joystick",
				PositionX:   posX,
				PositionY:   currentYPos,
				SortOrder:   (gIdx * 10) + cIdx + 2,
				QuizConfig:  &configStr,
				CreatedAt:   now,
			}
			if err := s.repos.CourseNode.Create(ctx, quizNode); err != nil {
				applog.Errorf("Failed to create quiz node %s: %v", quizNodeID, err)
			}

			// Connect from milestone or previous nodes
			connectFrom := prevNodeIDs
			if milestoneID != "" {
				connectFrom = []string{milestoneID}
			}

			for _, pID := range connectFrom {
				edgeCounter++
				eP := &model.CourseEdge{
					ID:           fmt.Sprintf("%s_e_%d", courseID, edgeCounter),
					CourseID:     courseID,
					SourceNodeID: pID,
					TargetNodeID: quizNodeID,
					EdgeType:     "required",
					CreatedAt:    now,
				}
				_ = s.repos.CourseEdge.Create(ctx, eP)
			}

			// Create the actual quiz & template
			quiz := &model.Quiz{
				ID:               actualQuizID,
				Title:            catTitle,
				Description:      fmt.Sprintf("Quiz: %s", catTitle),
				CourseID:         &courseID,
				NodeID:           &quizNodeID,
				DeckID:           &deck.ID,
				PassPercentage:   intPtr(70),
				ShuffleQuestions: true,
				QuestionMode:     "mixed",
				GivesCoins:       true,
				CoinReward:       utils.IfElse(group.QuizReward > 0, group.QuizReward, 10),
				IsPublic:         false,
				IsSystem:         true,
				IsDynamic:        true,
				CreatedBy:        createdBy,
				CreatedAt:        now,
				IsActive:         true,
			}
			if err := s.repos.Quiz.Create(ctx, quiz); err != nil {
				applog.Errorf("Failed to create quiz %d: %v", quiz.ID, err)
			}

			questTypes, _ := json.Marshal([]string{"mcq", "translate"})
			dirs, _ := json.Marshal([]string{"source_to_target", "target_to_source", "attach_suffix", "conjugate"})
			dID := deck.ID

			template := &model.QuestionTemplate{
				ID:             utils.GenerateSnowflakeID(),
				QuizID:         quiz.ID,
				DeckID:         &dID,
				CategoryID:     &catID,
				QuestionTypes:  string(questTypes),
				Directions:     string(dirs),
				GenerationMode: "random_from_deck",
				QuestionCount:  10,
				CreatedAt:      now,
			}
			if err := s.repos.QuestionTemplate.Create(ctx, template); err != nil {
				applog.Errorf("Failed to create template for quiz %d: %v", quiz.ID, err)
			}

			currentGroupEndNodes = append(currentGroupEndNodes, quizNodeID)
		}

		// E. Add synthèse quiz for multi-category groups
		if len(group.Categories) > 1 {
			currentYPos += 140.0

			synthNodeID := fmt.Sprintf("%s_synth_%d", courseID, gIdx)
			actualSynthQuizID := utils.GenerateSnowflakeID()

			synthQuestionCount := 10
			if len(group.Categories)*3 > 10 {
				synthQuestionCount = len(group.Categories) * 3
			}

			synthNodeConfig := map[string]interface{}{
				"deck_id":        deck.ID,
				"quiz_id":        strconv.FormatInt(actualSynthQuizID, 10),
				"question_types": []string{"mcq", "translate"},
				"directions":     []string{"source_to_target", "target_to_source", "attach_suffix", "conjugate"},
				"question_count": synthQuestionCount,
				"coin_reward":    utils.IfElse(group.SynthReward > 0, group.SynthReward, 30),
			}
			synthConfigJSON, _ := json.Marshal(synthNodeConfig)
			synthConfigStr := string(synthConfigJSON)

			synthNode := &model.CourseNode{
				ID:          synthNodeID,
				CourseID:    courseID,
				NodeType:    "quiz",
				Title:       "Synthèse",
				Description: fmt.Sprintf("Validation globale : %s", group.Name),
				Icon:        "sword",
				PositionX:   250,
				PositionY:   currentYPos,
				SortOrder:   (gIdx * 10) + 9,
				QuizConfig:  &synthConfigStr,
				CreatedAt:   now,
			}
			_ = s.repos.CourseNode.Create(ctx, synthNode)

			for cIdx, qNodeID := range currentGroupEndNodes {
				edgeCounter++
				e := &model.CourseEdge{
					ID:           fmt.Sprintf("%s_es_%d_%d", courseID, gIdx, cIdx),
					CourseID:     courseID,
					SourceNodeID: qNodeID,
					TargetNodeID: synthNodeID,
					EdgeType:     "required",
					CreatedAt:    now,
				}
				_ = s.repos.CourseEdge.Create(ctx, e)
			}

			// Synthèse Quiz
			synthQuiz := &model.Quiz{
				ID:               actualSynthQuizID,
				Title:            fmt.Sprintf("Synthèse : %s", group.Name),
				Description:      "Mettez vos connaissances à l'épreuve !",
				CourseID:         &courseID,
				NodeID:           &synthNodeID,
				DeckID:           &deck.ID,
				PassPercentage:   intPtr(80),
				ShuffleQuestions: true,
				QuestionMode:     "mixed",
				GivesCoins:       true,
				CoinReward:       utils.IfElse(group.SynthReward > 0, group.SynthReward, 30),
				IsPublic:         false,
				IsSystem:         true,
				IsDynamic:        true,
				CreatedBy:        createdBy,
				CreatedAt:        now,
				IsActive:         true,
			}
			_ = s.repos.Quiz.Create(ctx, synthQuiz)

			qPerCat := synthQuestionCount / len(group.Categories)
			if qPerCat < 2 {
				qPerCat = 2
			}
			questTypes, _ := json.Marshal([]string{"mcq", "translate"})
			dirs, _ := json.Marshal([]string{"source_to_target", "target_to_source", "attach_suffix", "conjugate"})

			for _, catKey := range group.Categories {
				catID, ok := categoryMap[catKey]
				if !ok {
					continue
				}
				dID := deck.ID
				template := &model.QuestionTemplate{
					ID:             utils.GenerateSnowflakeID(),
					QuizID:         synthQuiz.ID,
					DeckID:         &dID,
					CategoryID:     &catID,
					QuestionTypes:  string(questTypes),
					Directions:     string(dirs),
					GenerationMode: "random_from_deck",
					QuestionCount:  qPerCat,
					CreatedAt:      now,
				}
				_ = s.repos.QuestionTemplate.Create(ctx, template)
			}

			currentGroupEndNodes = []string{synthNodeID}
		}

		prevNodeIDs = currentGroupEndNodes
	}

	applog.Infof("Created course from template: %s with %d groups", course.Title, len(tmpl.Groups))
	return course, nil
}
