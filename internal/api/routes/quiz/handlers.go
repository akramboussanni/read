package quiz

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/akramboussanni/gocode/internal/applog"
	"github.com/akramboussanni/gocode/internal/model"
	quizsvc "github.com/akramboussanni/gocode/internal/quiz"
	"github.com/akramboussanni/gocode/internal/utils"
	"github.com/go-chi/chi/v5"
)

type OptionDTO struct {
	ID         string `json:"id"`
	OptionText string `json:"option_text"`
}

// QuestionDTO is a safe version of AttemptQuestion without the correct answer
type QuestionDTO struct {
	ID           string      `json:"id"`
	QuestionText string      `json:"question_text"`
	QuestionType string      `json:"question_type"`
	Direction    string      `json:"direction,omitempty"`
	Options      []OptionDTO `json:"options,omitempty"`
	DisplayOrder int         `json:"display_order"`
	Points       int         `json:"points"`
}

// ============================================================
// QUIZ CRUD
// ============================================================

// ListQuizzes returns public quizzes, paginated
func (qr *QuizRouter) ListQuizzes(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("page_size")
	offsetStr := r.URL.Query().Get("page")

	limit := 20
	offset := 0
	if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
		limit = v
	}
	if v, err := strconv.Atoi(offsetStr); err == nil && v > 0 {
		offset = (v - 1) * limit
	}

	quizzes, err := qr.Repos.Quiz.GetPublicQuizzes(r.Context(), limit, offset)
	if err != nil {
		quizzes = []*model.Quiz{}
	}

	response := map[string]interface{}{
		"quizzes":   quizzes,
		"total":     len(quizzes),
		"page":      offset/limit + 1,
		"page_size": limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateQuiz creates a standalone quiz
func (qr *QuizRouter) CreateQuiz(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	var body struct {
		Title           string `json:"title"`
		Description     string `json:"description"`
		IsPublic        bool   `json:"is_public"`
		IsDynamic       bool   `json:"is_dynamic"`
		PassPercentage  *int   `json:"pass_percentage"`
		GivesCoins      bool   `json:"gives_coins"`
		CoinReward      int    `json:"coin_reward"`
		ShuffleQuestion bool   `json:"shuffle_questions"`

		// Manual questions mode
		ManualQuestions []struct {
			QuestionText  string   `json:"question_text"`
			CorrectAnswer string   `json:"correct_answer"`
			Options       []string `json:"options"`
			QuestionType  string   `json:"question_type"`
			Direction     string   `json:"direction"`
		} `json:"manual_questions"`

		// Auto-generate mode
		AutoGenerate *struct {
			DeckSelections []struct {
				DeckID     string   `json:"deck_id"`
				Categories []string `json:"categories"`
			} `json:"deck_selections"`
			QuestionTypes []string `json:"question_types"`
			Directions    []string `json:"directions"`
			QuestionCount int      `json:"question_count"`
		} `json:"auto_generate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	now := utils.CurrentTimestamp()
	quizID := utils.GenerateSnowflakeID()

	// Determine question mode
	questionMode := "manual"
	isDynamic := body.IsDynamic
	if body.AutoGenerate != nil && len(body.AutoGenerate.DeckSelections) > 0 {
		questionMode = "auto"
		isDynamic = true
	}

	quiz := &model.Quiz{
		ID:               quizID,
		Title:            body.Title,
		Description:      body.Description,
		PassPercentage:   body.PassPercentage,
		ShuffleQuestions: body.ShuffleQuestion,
		QuestionMode:     questionMode,
		GivesCoins:       body.GivesCoins,
		CoinReward:       body.CoinReward,
		IsPublic:         body.IsPublic,
		IsDynamic:        isDynamic,
		CreatedBy:        &user.ID,
		CreatedAt:        now,
		IsActive:         true,
	}

	if body.PassPercentage == nil {
		pp := 70
		quiz.PassPercentage = &pp
	}

	if err := qr.Repos.Quiz.Create(r.Context(), quiz); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create quiz: %v", err), http.StatusInternalServerError)
		return
	}

	// Create question templates based on mode
	if body.AutoGenerate != nil && len(body.AutoGenerate.DeckSelections) > 0 {
		// Auto mode: create random_from_deck templates
		for _, sel := range body.AutoGenerate.DeckSelections {
			deckID, err := strconv.ParseInt(sel.DeckID, 10, 64)
			if err != nil {
				continue
			}

			questionTypes, _ := json.Marshal(body.AutoGenerate.QuestionTypes)
			directions, _ := json.Marshal(body.AutoGenerate.Directions)

			categories := sel.Categories

			if len(categories) == 0 {
				// If no categories passed, fetch all categories for this deck
				dbCats, err := qr.Repos.Category.GetByDeckID(r.Context(), deckID)
				if err == nil {
					for _, c := range dbCats {
						categories = append(categories, c.CategoryKey)
					}
				}
			}

			if len(categories) > 0 {
				// Per-category templates
				countPerCat := body.AutoGenerate.QuestionCount / len(categories)
				if countPerCat < 1 {
					countPerCat = 1
				}
				for _, catKey := range categories {
					cat, err := qr.Repos.Category.GetByKeyAndDeckID(r.Context(), deckID, catKey)
					if err != nil {
						continue
					}
					tmpl := &model.QuestionTemplate{
						ID:             utils.GenerateSnowflakeID(),
						QuizID:         quizID,
						DeckID:         &deckID,
						CategoryID:     &cat.ID, // Now this will use the implemented query
						QuestionTypes:  string(questionTypes),
						Directions:     string(directions),
						GenerationMode: "random_from_deck",
						QuestionCount:  countPerCat,
						CreatedAt:      now,
					}
					if err := qr.Repos.QuestionTemplate.Create(r.Context(), tmpl); err != nil {
						applog.Errorf("failed to save category template %s: %v", catKey, err)
					}
				}
			} else {
				// Fallback if deck has no categories somehow
				tmpl := &model.QuestionTemplate{
					ID:             utils.GenerateSnowflakeID(),
					QuizID:         quizID,
					DeckID:         &deckID,
					QuestionTypes:  string(questionTypes),
					Directions:     string(directions),
					GenerationMode: "random_from_deck",
					QuestionCount:  body.AutoGenerate.QuestionCount,
					CreatedAt:      now,
				}
				if err := qr.Repos.QuestionTemplate.Create(r.Context(), tmpl); err != nil {
					applog.Errorf("failed to save deck template: %v", err)
				}
			}
		}
	} else if len(body.ManualQuestions) > 0 {
		// Manual mode: create a single manual template with all questions
		manualData, _ := json.Marshal(body.ManualQuestions)
		manualDataStr := string(manualData)
		tmpl := &model.QuestionTemplate{
			ID:             utils.GenerateSnowflakeID(),
			QuizID:         quizID,
			QuestionTypes:  `["mcq","translate","write_word"]`,
			Directions:     `["source_to_target","target_to_source"]`,
			GenerationMode: "manual",
			ManualData:     &manualDataStr,
			QuestionCount:  len(body.ManualQuestions),
			CreatedAt:      now,
		}
		qr.Repos.QuestionTemplate.Create(r.Context(), tmpl)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(quiz)
}

// MyQuizzes returns quizzes created by the authenticated user
func (qr *QuizRouter) MyQuizzes(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	quizzes, err := qr.Repos.Quiz.GetQuizzesByCreator(r.Context(), user.ID)
	if err != nil || quizzes == nil {
		quizzes = []*model.Quiz{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quizzes)
}

// GetQuiz returns the quiz details and its templates
func (qr *QuizRouter) GetQuiz(w http.ResponseWriter, r *http.Request) {
	quizIDStr := chi.URLParam(r, "quizID")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid quiz ID", http.StatusBadRequest)
		return
	}

	quiz, err := qr.Repos.Quiz.GetByID(r.Context(), quizID)
	if err != nil {
		http.Error(w, "Quiz not found", http.StatusNotFound)
		return
	}

	templates, _ := qr.Repos.QuestionTemplate.GetByQuizID(r.Context(), quizID)

	response := map[string]interface{}{
		"quiz":      quiz,
		"templates": templates,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateQuiz updates the quiz configuration
func (qr *QuizRouter) UpdateQuiz(w http.ResponseWriter, r *http.Request) {
	quizIDStr := chi.URLParam(r, "quizID")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid quiz ID", http.StatusBadRequest)
		return
	}

	quiz, err := qr.Repos.Quiz.GetByID(r.Context(), quizID)
	if err != nil {
		http.Error(w, "Quiz not found", http.StatusNotFound)
		return
	}

	var body struct {
		Title            string `json:"title"`
		Description      string `json:"description"`
		PassPercentage   *int   `json:"pass_percentage"`
		ShuffleQuestions *bool  `json:"shuffle_questions"`
		QuestionMode     string `json:"question_mode"`
		IsPublic         *bool  `json:"is_public"`
		GivesCoins       *bool  `json:"gives_coins"`
		CoinReward       *int   `json:"coin_reward"`

		// Full template replacement
		Templates []*model.QuestionTemplate `json:"templates"`

		// Manual questions (convenience - replaces all manual templates)
		ManualQuestions []struct {
			QuestionText  string   `json:"question_text"`
			CorrectAnswer string   `json:"correct_answer"`
			Options       []string `json:"options"`
			QuestionType  string   `json:"question_type"`
			Direction     string   `json:"direction"`
		} `json:"manual_questions"`

		// Auto-generate config (convenience - replaces all auto templates)
		AutoGenerate *struct {
			DeckSelections []struct {
				DeckID     string   `json:"deck_id"`
				Categories []string `json:"categories"`
			} `json:"deck_selections"`
			QuestionTypes []string `json:"question_types"`
			Directions    []string `json:"directions"`
			QuestionCount int      `json:"question_count"`
		} `json:"auto_generate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Title != "" {
		quiz.Title = body.Title
	}
	if body.Description != "" {
		quiz.Description = body.Description
	}
	if body.PassPercentage != nil {
		quiz.PassPercentage = body.PassPercentage
	}
	if body.ShuffleQuestions != nil {
		quiz.ShuffleQuestions = *body.ShuffleQuestions
	}
	if body.QuestionMode != "" {
		quiz.QuestionMode = body.QuestionMode
	}
	if body.IsPublic != nil {
		quiz.IsPublic = *body.IsPublic
	}
	if body.GivesCoins != nil {
		quiz.GivesCoins = *body.GivesCoins
	}
	if body.CoinReward != nil {
		quiz.CoinReward = *body.CoinReward
	}

	if err := qr.Repos.Quiz.UpdateQuiz(r.Context(), quiz); err != nil {
		http.Error(w, "Failed to update quiz", http.StatusInternalServerError)
		return
	}

	now := utils.CurrentTimestamp()

	// Handle template updates
	if len(body.ManualQuestions) > 0 {
		// Replace all templates with a single manual template
		qr.Repos.QuestionTemplate.DeleteByQuizID(r.Context(), quizID)
		manualData, _ := json.Marshal(body.ManualQuestions)
		manualDataStr := string(manualData)
		tmpl := &model.QuestionTemplate{
			ID:             utils.GenerateSnowflakeID(),
			QuizID:         quizID,
			QuestionTypes:  `["mcq","translate","write_word"]`,
			Directions:     `["source_to_target","target_to_source"]`,
			GenerationMode: "manual",
			ManualData:     &manualDataStr,
			QuestionCount:  len(body.ManualQuestions),
			CreatedAt:      now,
		}
		qr.Repos.QuestionTemplate.Create(r.Context(), tmpl)
	} else if body.AutoGenerate != nil {
		// Replace all templates with auto-generate templates
		qr.Repos.QuestionTemplate.DeleteByQuizID(r.Context(), quizID)
		for _, sel := range body.AutoGenerate.DeckSelections {
			deckID, err := strconv.ParseInt(sel.DeckID, 10, 64)
			if err != nil {
				continue
			}
			questionTypes, _ := json.Marshal(body.AutoGenerate.QuestionTypes)
			directions, _ := json.Marshal(body.AutoGenerate.Directions)

			if len(sel.Categories) == 0 {
				tmpl := &model.QuestionTemplate{
					ID:             utils.GenerateSnowflakeID(),
					QuizID:         quizID,
					DeckID:         &deckID,
					QuestionTypes:  string(questionTypes),
					Directions:     string(directions),
					GenerationMode: "random_from_deck",
					QuestionCount:  body.AutoGenerate.QuestionCount,
					CreatedAt:      now,
				}
				qr.Repos.QuestionTemplate.Create(r.Context(), tmpl)
			} else {
				countPerCat := body.AutoGenerate.QuestionCount / len(sel.Categories)
				if countPerCat < 1 {
					countPerCat = 1
				}
				for _, catKey := range sel.Categories {
					cat, err := qr.Repos.Category.GetByKeyAndDeckID(r.Context(), deckID, catKey)
					if err != nil {
						continue
					}
					tmpl := &model.QuestionTemplate{
						ID:             utils.GenerateSnowflakeID(),
						QuizID:         quizID,
						DeckID:         &deckID,
						CategoryID:     &cat.ID,
						QuestionTypes:  string(questionTypes),
						Directions:     string(directions),
						GenerationMode: "random_from_deck",
						QuestionCount:  countPerCat,
						CreatedAt:      now,
					}
					qr.Repos.QuestionTemplate.Create(r.Context(), tmpl)
				}
			}
		}
	} else if len(body.Templates) > 0 {
		// Direct template update (raw)
		for _, tmpl := range body.Templates {
			qr.Repos.QuestionTemplate.Update(r.Context(), tmpl)
		}
	}

	// Return updated quiz with templates
	templates, _ := qr.Repos.QuestionTemplate.GetByQuizID(r.Context(), quizID)
	response := map[string]interface{}{
		"quiz":      quiz,
		"templates": templates,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteQuiz soft-deletes a quiz
func (qr *QuizRouter) DeleteQuiz(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())
	quizIDStr := chi.URLParam(r, "quizID")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid quiz ID", http.StatusBadRequest)
		return
	}

	quiz, err := qr.Repos.Quiz.GetByID(r.Context(), quizID)
	if err != nil {
		http.Error(w, "Quiz not found", http.StatusNotFound)
		return
	}

	// Only creator or admin can delete
	if quiz.CreatedBy != nil && *quiz.CreatedBy != user.ID && !user.IsAdmin {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if err := qr.Repos.Quiz.DeactivateQuiz(r.Context(), quizID); err != nil {
		http.Error(w, "Failed to delete quiz", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================
// DECK BROWSING
// ============================================================

// ListDecks returns all available decks with entry counts
func (qr *QuizRouter) ListDecks(w http.ResponseWriter, r *http.Request) {
	decks, err := qr.Repos.Deck.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to get decks", http.StatusInternalServerError)
		return
	}

	type DeckWithCounts struct {
		*model.Deck
		CategoryCount int `json:"category_count"`
		QuestionCount int `json:"question_count"`
	}

	result := make([]DeckWithCounts, 0, len(decks))
	for _, deck := range decks {
		catCount, _ := qr.Repos.Category.CountByDeckID(r.Context(), deck.ID)
		// Rough question count estimate
		entries, _ := qr.Repos.DeckEntry.GetByDeckID(r.Context(), deck.ID)
		result = append(result, DeckWithCounts{
			Deck:          deck,
			CategoryCount: catCount,
			QuestionCount: len(entries),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetDeckCategories returns categories for a specific deck
func (qr *QuizRouter) GetDeckCategories(w http.ResponseWriter, r *http.Request) {
	deckIDStr := chi.URLParam(r, "deckID")
	deckID, err := strconv.ParseInt(deckIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid deck ID", http.StatusBadRequest)
		return
	}

	categories, err := qr.Repos.Category.GetByDeckID(r.Context(), deckID)
	if err != nil {
		http.Error(w, "Failed to get categories", http.StatusInternalServerError)
		return
	}

	// Include entry counts
	type CategoryWithCount struct {
		*model.Category
		QuestionCount int `json:"question_count"`
	}

	result := make([]CategoryWithCount, 0, len(categories))
	for _, cat := range categories {
		count, _ := qr.Repos.DeckEntry.CountByCategoryID(r.Context(), cat.ID)
		result = append(result, CategoryWithCount{
			Category:      cat,
			QuestionCount: count,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ============================================================
// QUESTION PREVIEW
// ============================================================

// PreviewQuestions generates sample questions for manual review
func (qr *QuizRouter) PreviewQuestions(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DeckSelections []quizsvc.PreviewDeckSelection `json:"deck_selections"`
		QuestionTypes  []string                       `json:"question_types"`
		Directions     []string                       `json:"directions"`
		QuestionCount  int                            `json:"question_count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.QuestionCount < 1 {
		body.QuestionCount = 1
	}
	if body.QuestionCount > 50 {
		body.QuestionCount = 50
	}

	questions, err := qr.QuizService.PreviewQuestions(r.Context(), body.DeckSelections, body.QuestionTypes, body.Directions, body.QuestionCount)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate preview: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions)
}

// ============================================================
// QUIZ EXECUTION
// ============================================================

// StartQuiz starts a quiz attempt
func (qr *QuizRouter) StartQuiz(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	var body struct {
		QuizID       int64   `json:"quiz_id,string"`
		CourseID     *string `json:"course_id,omitempty"`
		NodeID       *string `json:"node_id,omitempty"`
		AssignmentID *int64  `json:"assignment_id,string,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	attempt, questions, err := qr.QuizService.StartQuiz(r.Context(), user.ID, body.QuizID, body.CourseID, body.NodeID, body.AssignmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert questions to safe DTOs (no correct answers)
	dtos := make([]QuestionDTO, len(questions))
	for i, q := range questions {
		var rawOpts []string
		var optDTOs []OptionDTO
		if q.Options != "" && q.Options != "null" {
			json.Unmarshal([]byte(q.Options), &rawOpts)
			optDTOs = make([]OptionDTO, len(rawOpts))
			for j, optText := range rawOpts {
				optDTOs[j] = OptionDTO{
					ID:         strconv.Itoa(j + 1), // Simple ID based on index
					OptionText: optText,
				}
			}
		}
		dir := ""
		if q.Direction != nil {
			dir = *q.Direction
		}
		dtos[i] = QuestionDTO{
			ID:           strconv.FormatInt(q.ID, 10),
			QuestionText: q.QuestionText,
			QuestionType: q.QuestionType,
			Direction:    dir,
			Options:      optDTOs,
			DisplayOrder: q.DisplayOrder,
			Points:       1,
		}
	}

	// Get previous answers if resuming
	// Join with attempt_questions to get the correct answer for the feedback display
	var prevResponses []struct {
		QuestionID    int64   `db:"question_id"`
		UserAnswer    string  `db:"user_answer"`
		IsCorrect     bool    `db:"is_correct"`
		CorrectAnswer string  `db:"correct_answer"`
		AIExplanation *string `db:"ai_explanation"`
		PointsEarned  float64 `db:"points_earned"`
	}
	err = qr.Repos.Quiz.GetDB().SelectContext(r.Context(), &prevResponses, `
		SELECT ua.question_id, ua.user_answer, ua.is_correct, aq.correct_answer, ua.ai_explanation, ua.points_earned
		FROM user_answers ua
		JOIN attempt_questions aq ON ua.question_id = aq.id
		WHERE ua.attempt_id = $1
		ORDER BY ua.answered_at
	`, attempt.ID)

	var prevAnswerDTOs []map[string]interface{}
	if err == nil {
		for _, a := range prevResponses {
			dto := map[string]interface{}{
				"question_id":    strconv.FormatInt(a.QuestionID, 10),
				"answer":         a.UserAnswer,
				"is_correct":     a.IsCorrect,
				"correct_answer": a.CorrectAnswer,
				"points_earned":  a.PointsEarned,
			}
			if a.AIExplanation != nil {
				dto["ai_explanation"] = *a.AIExplanation
			}
			prevAnswerDTOs = append(prevAnswerDTOs, dto)
		}
	}

	response := map[string]interface{}{
		"attempt_id":       strconv.FormatInt(attempt.ID, 10),
		"quiz_id":          strconv.FormatInt(attempt.QuizID, 10),
		"questions":        dtos,
		"previous_answers": prevAnswerDTOs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SubmitAnswer submits an answer for a single question
func (qr *QuizRouter) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	var body struct {
		AttemptID  int64  `json:"attempt_id,string"`
		QuestionID int64  `json:"question_id,string"`
		Answer     string `json:"answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	answer, needsMoreDetail, err := qr.QuizService.SubmitAnswer(r.Context(), user.ID, body.AttemptID, body.QuestionID, body.Answer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the question for the correct answer
	question, _ := qr.Repos.AttemptQuestion.GetByID(r.Context(), body.QuestionID)

	response := map[string]interface{}{
		"is_correct":        answer.IsCorrect,
		"points_earned":     answer.PointsEarned,
		"correct_answer":    "",
		"needs_more_detail": needsMoreDetail,
	}
	if question != nil && !needsMoreDetail {
		// Only reveal the correct answer once the question is actually locked in (not on a warning)
		response["correct_answer"] = question.CorrectAnswer
	}
	if answer.AIExplanation != nil {
		response["ai_explanation"] = *answer.AIExplanation
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CompleteQuiz completes a quiz attempt and returns results
func (qr *QuizRouter) CompleteQuiz(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	var body struct {
		AttemptID int64 `json:"attempt_id,string"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := qr.QuizService.CompleteQuiz(r.Context(), user.ID, body.AttemptID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetAttempt returns details of a completed attempt
func (qr *QuizRouter) GetAttempt(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())
	attemptIDStr := chi.URLParam(r, "attemptID")
	attemptID, err := strconv.ParseInt(attemptIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid attempt ID", http.StatusBadRequest)
		return
	}

	attempt, err := qr.Repos.QuizAttempt.GetByID(r.Context(), attemptID)
	if err != nil {
		http.Error(w, "Attempt not found", http.StatusNotFound)
		return
	}
	if attempt.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	questions, _ := qr.Repos.AttemptQuestion.GetByAttemptID(r.Context(), attemptID)
	answers, _ := qr.Repos.UserAnswer.GetByAttemptID(r.Context(), attemptID)

	response := map[string]interface{}{
		"attempt":   attempt,
		"questions": questions,
		"answers":   answers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MyHistory returns the user's quiz history
func (qr *QuizRouter) MyHistory(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}

	attempts, err := qr.Repos.QuizAttempt.GetUserAttempts(r.Context(), user.ID, limit)
	if err != nil {
		http.Error(w, "Failed to get history", http.StatusInternalServerError)
		return
	}

	// Enrich with quiz titles
	type AttemptWithTitle struct {
		*model.QuizAttempt
		QuizTitle string `json:"quiz_title"`
	}

	var result []AttemptWithTitle
	for _, a := range attempts {
		quizTitle := fmt.Sprintf("Quiz %d", a.QuizID)
		quiz, err := qr.Repos.Quiz.GetByID(r.Context(), a.QuizID)
		if err == nil {
			quizTitle = quiz.Title
		}
		result = append(result, AttemptWithTitle{
			QuizAttempt: a,
			QuizTitle:   quizTitle,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// CompleteNode marks a generic learning node as completed for the current user.
func (qr *QuizRouter) CompleteNode(w http.ResponseWriter, r *http.Request) {
	user, _ := utils.UserFromContext(r.Context())

	var body struct {
		CourseID string `json:"course_id"`
		NodeID   string `json:"node_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := qr.QuizService.CompleteNode(r.Context(), user.ID, body.CourseID, body.NodeID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
