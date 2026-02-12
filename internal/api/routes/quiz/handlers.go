package quiz

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/akramboussanni/gocode/internal/api"
	"github.com/akramboussanni/gocode/internal/applog"
	"github.com/akramboussanni/gocode/internal/model"
	quizpkg "github.com/akramboussanni/gocode/internal/quiz"
	"github.com/akramboussanni/gocode/internal/utils"
	"github.com/go-chi/chi/v5"
)

func (qr *QuizRouter) HandleGetQuiz(w http.ResponseWriter, r *http.Request) {
	quizIDStr := chi.URLParam(r, "quizID")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid quiz ID")
		return
	}

	// Get quiz from database
	quizModel, err := qr.Repos.Quiz.GetByID(r.Context(), quizID)
	if err != nil {
		applog.Error("Failed to get quiz:", err)
		api.WriteNotFound(w, "Quiz not found")
		return
	}

	// Get quiz questions
	quizQuestions, err := qr.Repos.QuizQuestion.GetByQuizID(r.Context(), quizID)
	if err != nil {
		applog.Error("Failed to get quiz questions:", err)
		api.WriteInternalError(w)
		return
	}

	// Get deck selections
	deckSelections, err := qr.Repos.QuizCategorySelection.GetByQuizID(r.Context(), quizID)
	if err != nil {
		applog.Error("Failed to get deck selections:", err)
		api.WriteInternalError(w)
		return
	}

	// Convert to API response
	var questions []QuizQuestionResponse
	for _, q := range quizQuestions {
		var options []string
		if q.Options != "" && q.Options != "[]" {
			json.Unmarshal([]byte(q.Options), &options)
		}

		direction := ""
		if q.Direction != nil {
			direction = *q.Direction
		}

		questions = append(questions, QuizQuestionResponse{
			ID:           q.ID,
			QuestionText: q.QuestionText,
			Options:      options,
			QuestionType: q.QuestionType,
			Direction:    direction,
		})
	}

	// Build config from deck selections (simplified)
	var deckSelectionsAPI []quizpkg.DeckSelection
	// This is a simplified version - in reality we'd need to group by deck
	for _, sel := range deckSelections {
		// Get category info
		category, err := qr.Repos.Category.GetByID(r.Context(), sel.CategoryID)
		if err != nil {
			continue
		}
		deckSelectionsAPI = append(deckSelectionsAPI, quizpkg.DeckSelection{
			DeckID:     category.DeckID,
			Categories: []string{category.CategoryKey},
		})
	}

	config := quizpkg.QuizConfig{
		DeckSelections: deckSelectionsAPI,
		QuestionCount:  len(questions),
	}

	response := QuizDetailResponse{
		ID:           quizModel.ID,
		Title:        quizModel.Title,
		Description:  quizModel.Description,
		Config:       config,
		Questions:    questions,
		CurrentIndex: 0, // Not tracking progress yet
		Progress: QuizProgress{
			Answered:   0,
			Total:      len(questions),
			Percentage: 0,
		},
	}

	api.WriteJSON(w, 200, response)
}

func (qr *QuizRouter) HandleListQuizzes(w http.ResponseWriter, r *http.Request) {
	// Get public quizzes
	quizzes, err := qr.Repos.Quiz.GetPublicQuizzes(r.Context(), 50, 0) // Limit 50 for now
	if err != nil {
		applog.Error("Failed to get quizzes:", err)
		api.WriteInternalError(w)
		return
	}

	var response []QuizListItem
	for _, quiz := range quizzes {
		// Count questions for this quiz
		questions, err := qr.Repos.QuizQuestion.GetByQuizID(r.Context(), quiz.ID)
		if err != nil {
			applog.Error("Failed to count questions for quiz %d: %v", quiz.ID, err)
			continue
		}

		creatorName := ""
		if quiz.CreatedBy != nil {
			// Get user name
			user, err := qr.Repos.User.GetUserByID(r.Context(), *quiz.CreatedBy)
			if err == nil {
				creatorName = user.Username
			}
		}

		response = append(response, QuizListItem{
			ID:            quiz.ID,
			Title:         quiz.Title,
			Description:   quiz.Description,
			CreatorName:   creatorName,
			QuestionCount: len(questions),
			IsPublic:      quiz.IsPublic,
			CreatedAt:     time.Unix(quiz.CreatedAt, 0),
		})
	}

	api.WriteJSON(w, 200, response)
}

func (qr *QuizRouter) HandleCreateQuiz(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}

	req, err := api.DecodeJSON[CreateQuizRequest](w, r)
	if err != nil {
		applog.Error("Failed to decode quiz request:", err)
		return
	}

	// Validate basic requirements
	if req.Title == "" {
		api.WriteBadRequest(w, "Title is required")
		return
	}

	if len(req.ManualQuestions) == 0 && req.AutoGenerate == nil {
		api.WriteBadRequest(w, "Either manual questions or auto-generation config is required")
		return
	}

	// Validate admin-only fields
	if req.PassPercentage != nil && !user.IsAdmin {
		api.WriteError(w, http.StatusForbidden, "Only admins can set pass percentage")
		return
	}
	if req.GivesCoins && !user.IsAdmin {
		api.WriteError(w, http.StatusForbidden, "Only admins can enable coin rewards")
		return
	}
	if req.CoinReward > 0 && !user.IsAdmin {
		api.WriteError(w, http.StatusForbidden, "Only admins can set coin rewards")
		return
	}
	if req.LevelOrder > 0 && !user.IsAdmin {
		api.WriteError(w, http.StatusForbidden, "Only admins can set level order")
		return
	}
	if req.PrerequisiteQuizID != nil && !user.IsAdmin {
		api.WriteError(w, http.StatusForbidden, "Only admins can set prerequisite quizzes")
		return
	}
	if req.IsSystem && !user.IsAdmin {
		api.WriteError(w, http.StatusForbidden, "Only admins can create system quizzes")
		return
	}

	// Start building the quiz questions
	var allQuestions []quizpkg.QuizQuestion

	// Process manual questions
	for _, mq := range req.ManualQuestions {
		question := quizpkg.QuizQuestion{
			QuestionText:  mq.QuestionText,
			CorrectAnswer: mq.CorrectAnswer,
			Options:       mq.Options,
		}

		// Set question type
		switch mq.QuestionType {
		case "mcq":
			question.QuestionType = quizpkg.QuestionTypeMCQ
		case "write_word":
			question.QuestionType = quizpkg.QuestionTypeWriteWord
		case "translate":
			question.QuestionType = quizpkg.QuestionTypeTranslate
		default:
			api.WriteBadRequest(w, "Invalid question type: "+mq.QuestionType)
			return
		}

		// Set direction
		switch mq.Direction {
		case "source_to_target":
			question.Direction = quizpkg.DirectionSourceToTarget
		case "target_to_source":
			question.Direction = quizpkg.DirectionTargetToSource
		default:
			api.WriteBadRequest(w, "Invalid direction: "+mq.Direction)
			return
		}

		allQuestions = append(allQuestions, question)
	}

	// Process auto-generated questions if specified
	if req.AutoGenerate != nil {
		// Load decks for auto-generation
		var parsedDecks []*quizpkg.ParsedDeck
		deckMap := make(map[int64]*quizpkg.ParsedDeck)

		for _, sel := range req.AutoGenerate.DeckSelections {
			if _, exists := deckMap[sel.DeckID]; !exists {
				deck, err := qr.DeckService.GetDeck(r.Context(), sel.DeckID)
				if err != nil {
					applog.Error("Failed to load deck %d: %v", sel.DeckID, err)
					api.WriteInternalError(w)
					return
				}
				parsedDecks = append(parsedDecks, deck)
				deckMap[sel.DeckID] = deck
			}
		}

		// Convert auto-generate config to service types
		var deckSelections []quizpkg.DeckSelection
		var questionTypes []quizpkg.QuestionType
		var directions []quizpkg.Direction

		for _, sel := range req.AutoGenerate.DeckSelections {
			deckSelections = append(deckSelections, quizpkg.DeckSelection{
				DeckID:     sel.DeckID,
				Categories: sel.Categories,
			})
		}

		for _, qt := range req.AutoGenerate.QuestionTypes {
			switch qt {
			case "mcq":
				questionTypes = append(questionTypes, quizpkg.QuestionTypeMCQ)
			case "write_word":
				questionTypes = append(questionTypes, quizpkg.QuestionTypeWriteWord)
			case "translate":
				questionTypes = append(questionTypes, quizpkg.QuestionTypeTranslate)
			}
		}

		for _, dir := range req.AutoGenerate.Directions {
			switch dir {
			case "source_to_target":
				directions = append(directions, quizpkg.DirectionSourceToTarget)
			case "target_to_source":
				directions = append(directions, quizpkg.DirectionTargetToSource)
			}
		}

		config := quizpkg.QuizConfig{
			DeckSelections:  deckSelections,
			QuestionTypes:   questionTypes,
			Directions:      directions,
			QuestionCount:   req.AutoGenerate.QuestionCount,
			Difficulty:      req.AutoGenerate.Difficulty,
			CustomQuestions: nil, // No custom questions for auto-generation
		}

		autoQuiz, err := qr.QuizService.CreateQuiz(r.Context(), config, parsedDecks)
		if err != nil {
			applog.Error("Failed to auto-generate quiz questions:", err)
			api.WriteInternalError(w)
			return
		}

		// Add auto-generated questions to the list
		allQuestions = append(allQuestions, autoQuiz.Questions...)
	}

	// Create a quiz object with all questions
	// Note: We don't need to create a Quiz object here since we're just storing the questions
	// The Quiz struct is for runtime quiz state, not for creation

	// Generate quiz ID
	quizID := utils.GenerateSnowflakeID()

	// Create quiz record
	quizModel := &model.Quiz{
		ID:                 quizID,
		Title:              req.Title,
		Description:        req.Description,
		Version:            1,   // Start at version 1
		DeckID:             nil, // Multi-deck, so no single deck
		TimeLimit:          nil, // Always nil as requested
		PassPercentage:     req.PassPercentage,
		ShuffleQuestions:   true,
		QuestionMode:       "source_to_target",
		GivesCoins:         req.GivesCoins,
		CoinReward:         req.CoinReward,
		LevelOrder:         req.LevelOrder,
		PrerequisiteQuizID: req.PrerequisiteQuizID,
		IsPublic:           req.IsPublic,
		IsSystem:           req.IsSystem,
		CreatedBy:          &user.ID,
		CreatedAt:          time.Now().Unix(),
		IsActive:           true,
	}

	// Begin transaction
	tx, err := qr.Repos.Quiz.BeginTx(r.Context())
	if err != nil {
		applog.Error("Failed to begin transaction:", err)
		api.WriteInternalError(w)
		return
	}
	defer tx.Rollback()

	// Create quiz
	if err := qr.Repos.Quiz.CreateWithTx(r.Context(), tx, quizModel); err != nil {
		applog.Error("Failed to create quiz:", err)
		api.WriteInternalError(w)
		return
	}

	// Store deck selections if auto-generation was used
	if req.AutoGenerate != nil {
		for _, sel := range req.AutoGenerate.DeckSelections {
			// Get categories for this deck
			categories, err := qr.Repos.Category.GetByDeckID(r.Context(), sel.DeckID)
			if err != nil {
				applog.Error("Failed to get categories for deck %d: %v", sel.DeckID, err)
				api.WriteInternalError(w)
				return
			}

			// If specific categories selected, only include those
			if len(sel.Categories) > 0 {
				filteredCategories := make([]*model.Category, 0)
				for _, cat := range categories {
					for _, selCat := range sel.Categories {
						if cat.CategoryKey == selCat {
							filteredCategories = append(filteredCategories, cat)
							break
						}
					}
				}
				categories = filteredCategories
			}

			// Calculate question count per category
			questionCount := req.AutoGenerate.QuestionCount / len(categories)
			if questionCount == 0 {
				questionCount = 1
			}

			for _, cat := range categories {
				selection := &model.QuizCategorySelection{
					QuizID:        quizID,
					CategoryID:    cat.ID,
					QuestionCount: questionCount,
				}
				if err := qr.Repos.QuizCategorySelection.CreateWithTx(r.Context(), tx, selection); err != nil {
					applog.Error("Failed to create quiz category selection:", err)
					api.WriteInternalError(w)
					return
				}
			}
		}
	}

	// Store quiz questions
	for i, q := range allQuestions {
		// Convert options to JSON
		optionsJSON := "[]"
		if len(q.Options) > 0 {
			if jsonData, err := json.Marshal(q.Options); err == nil {
				optionsJSON = string(jsonData)
			}
		}

		direction := ""
		if q.Direction != "" {
			direction = string(q.Direction)
		}

		quizQuestion := &model.QuizQuestion{
			ID:            utils.GenerateSnowflakeID(),
			QuizID:        quizID,
			QuestionID:    nil, // For now, not linking to original questions
			QuestionText:  q.QuestionText,
			CorrectAnswer: q.CorrectAnswer,
			Options:       optionsJSON,
			QuestionType:  string(q.QuestionType),
			Direction:     &direction,
			DisplayOrder:  i + 1,
			CreatedAt:     time.Now().Unix(),
		}

		if err := qr.Repos.QuizQuestion.CreateWithTx(r.Context(), tx, quizQuestion); err != nil {
			applog.Error("Failed to create quiz question:", err)
			api.WriteInternalError(w)
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		applog.Error("Failed to commit transaction:", err)
		api.WriteInternalError(w)
		return
	}

	// Convert to API response
	var questions []QuizQuestionResponse
	for _, q := range allQuestions {
		questions = append(questions, QuizQuestionResponse{
			ID:           q.ID,
			QuestionText: q.QuestionText,
			Options:      q.Options,
			QuestionType: string(q.QuestionType),
			Direction:    string(q.Direction),
		})
	}

	response := QuizDetailResponse{
		ID:           quizID,
		Title:        req.Title,
		Description:  req.Description,
		Config:       quizpkg.QuizConfig{}, // Empty config for response
		Questions:    questions,
		CurrentIndex: 0,
		StartedAt:    nil,
		Progress: QuizProgress{
			Answered:   0,
			Total:      len(allQuestions),
			Percentage: 0,
		},
	}

	api.WriteJSON(w, 201, response)
}

func (qr *QuizRouter) HandleUpdateQuiz(w http.ResponseWriter, r *http.Request) {
	// For the new system, quizzes are not stored persistently
	api.WriteMessage(w, 501, "error", "Quiz updates not supported in the new system")
}

func (qr *QuizRouter) HandleDeleteQuiz(w http.ResponseWriter, r *http.Request) {
	// For the new system, quizzes are not stored persistently
	api.WriteMessage(w, 501, "error", "Quiz deletion not supported in the new system")
}

func (qr *QuizRouter) HandleStartQuiz(w http.ResponseWriter, r *http.Request) {
	// For the new system, quizzes start when created
	api.WriteMessage(w, 501, "error", "Quiz starting not applicable in the new system")
}

func (qr *QuizRouter) HandleSubmitQuiz(w http.ResponseWriter, r *http.Request) {
	// For the new system, answers are submitted individually
	api.WriteMessage(w, 501, "error", "Use individual answer submission endpoints")
}

func (qr *QuizRouter) HandleCreateQuestion(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	_ = user

	req, err := api.DecodeJSON[CreateQuestionRequest](w, r)
	if err != nil {
		applog.Error("Failed to decode question request:", err)
		return
	}

	if req.QuestionText == "" || req.CorrectAnswer == "" {
		api.WriteBadRequest(w, "Question text and correct answer are required")
		return
	}

	// For now, just return success - custom question creation not fully implemented
	api.WriteMessage(w, 201, "success", "Question created successfully")
}

func (qr *QuizRouter) HandleListDecks(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	_ = user

	// Get all decks from database
	decks, err := qr.Repos.Deck.GetAll(r.Context())
	if err != nil {
		applog.Error("Failed to get decks:", err)
		api.WriteInternalError(w)
		return
	}

	response := make([]DeckListItem, 0, len(decks))
	for _, deck := range decks {
		// Get category count for this deck
		categories, err := qr.Repos.Category.GetByDeckID(r.Context(), deck.ID)
		if err != nil {
			applog.Warn("Failed to get categories for deck %d: %v", deck.ID, err)
			continue
		}

		// Get question count for this deck
		entries, err := qr.Repos.DeckEntry.GetByDeckID(r.Context(), deck.ID)
		if err != nil {
			applog.Warn("Failed to get entries for deck %d: %v", deck.ID, err)
			continue
		}

		response = append(response, DeckListItem{
			ID:            deck.ID,
			DeckKey:       deck.DeckKey,
			Title:         deck.Title,
			CategoryCount: len(categories),
			QuestionCount: len(entries),
		})
	}

	api.WriteJSON(w, 200, response)
}

func (qr *QuizRouter) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	_ = user

	deckID, err := strconv.ParseInt(chi.URLParam(r, "deckID"), 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid deck ID")
		return
	}

	// Get categories for the specified deck
	categories, err := qr.Repos.Category.GetByDeckID(r.Context(), deckID)
	if err != nil {
		applog.Error("Failed to get categories for deck %d: %v", deckID, err)
		api.WriteInternalError(w)
		return
	}

	response := make([]CategoryListItem, 0, len(categories))
	for _, category := range categories {
		// Count entries in this category
		entries, err := qr.Repos.DeckEntry.GetByDeckAndCategoryID(r.Context(), deckID, category.ID)
		if err != nil {
			applog.Warn("Failed to get entries for category %d: %v", category.ID, err)
			continue
		}

		response = append(response, CategoryListItem{
			ID:            category.ID,
			CategoryKey:   category.CategoryKey,
			Title:         category.Title,
			QuestionCount: len(entries),
		})
	}

	api.WriteJSON(w, 200, response)
}

func (qr *QuizRouter) HandleSubmitAnswer(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}

	quizIDStr := chi.URLParam(r, "quizID")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid quiz ID")
		return
	}

	req, err := api.DecodeJSON[SubmitAnswerRequest](w, r)
	if err != nil {
		applog.Error("Failed to decode answer request:", err)
		return
	}

	// Get quiz question
	quizQuestion, err := qr.Repos.QuizQuestion.GetByID(r.Context(), req.QuestionID)
	if err != nil {
		applog.Error("Failed to get quiz question:", err)
		api.WriteNotFound(w, "Question not found")
		return
	}

	if quizQuestion.QuizID != quizID {
		api.WriteBadRequest(w, "Question does not belong to this quiz")
		return
	}

	// Check if attempt exists, create if not
	attempt, err := qr.Repos.QuizAttempt.GetByUserAndQuiz(r.Context(), user.ID, quizID)
	if err != nil {
		// Create new attempt
		attemptID := utils.GenerateSnowflakeID()
		attempt = &model.QuizAttempt{
			ID:          attemptID,
			UserID:      user.ID,
			QuizID:      quizID,
			StartedAt:   time.Now().Unix(),
			CoinsEarned: 0,
		}
		if err := qr.Repos.QuizAttempt.Create(r.Context(), attempt); err != nil {
			applog.Error("Failed to create quiz attempt:", err)
			api.WriteInternalError(w)
			return
		}
	}

	// Check if answer already exists
	existingAnswer, err := qr.Repos.UserAnswer.GetByAttemptAndQuestion(r.Context(), attempt.ID, req.QuestionID)
	if err == nil && existingAnswer != nil {
		api.WriteBadRequest(w, "Answer already submitted for this question")
		return
	}

	// Validate the answer
	isCorrect := strings.ToLower(strings.TrimSpace(req.Answer)) == strings.ToLower(strings.TrimSpace(quizQuestion.CorrectAnswer))

	// Calculate points (simple: 1 for correct, 0 for incorrect)
	pointsEarned := 0.0
	if isCorrect {
		pointsEarned = 1.0
	}

	// Create user answer
	userAnswer := &model.UserAnswer{
		ID:           utils.GenerateSnowflakeID(),
		AttemptID:    attempt.ID,
		QuestionID:   req.QuestionID,
		UserAnswer:   req.Answer,
		IsCorrect:    isCorrect,
		PointsEarned: pointsEarned,
		AnsweredAt:   time.Now().Unix(),
	}

	if err := qr.Repos.UserAnswer.Create(r.Context(), userAnswer); err != nil {
		applog.Error("Failed to create user answer:", err)
		api.WriteInternalError(w)
		return
	}

	// Calculate progress
	totalQuestions, err := qr.Repos.QuizQuestion.CountByQuizID(r.Context(), quizID)
	if err != nil {
		applog.Error("Failed to count questions:", err)
		api.WriteInternalError(w)
		return
	}

	answeredCount, err := qr.Repos.UserAnswer.CountByAttemptID(r.Context(), attempt.ID)
	if err != nil {
		applog.Error("Failed to count answers:", err)
		api.WriteInternalError(w)
		return
	}

	correctCount, err := qr.Repos.UserAnswer.CountCorrectByAttemptID(r.Context(), attempt.ID)
	if err != nil {
		applog.Error("Failed to count correct answers:", err)
		api.WriteInternalError(w)
		return
	}

	percentage := float64(answeredCount) / float64(totalQuestions) * 100
	score := float64(correctCount) / float64(totalQuestions) * 100

	response := SubmitAnswerResponse{
		IsCorrect:     isCorrect,
		CorrectAnswer: quizQuestion.CorrectAnswer,
		Progress: QuizProgress{
			Answered:   answeredCount,
			Total:      totalQuestions,
			Percentage: percentage,
			Correct:    correctCount,
			Score:      score,
		},
	}

	api.WriteJSON(w, 200, response)
}

func (qr *QuizRouter) HandleGetQuizProgress(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}

	quizIDStr := chi.URLParam(r, "quizID")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid quiz ID")
		return
	}

	// Get or create attempt
	attempt, err := qr.Repos.QuizAttempt.GetByUserAndQuiz(r.Context(), user.ID, quizID)
	if err != nil {
		// No attempt yet, return zero progress
		totalQuestions, err := qr.Repos.QuizQuestion.CountByQuizID(r.Context(), quizID)
		if err != nil {
			applog.Error("Failed to count questions:", err)
			api.WriteInternalError(w)
			return
		}

		response := QuizProgress{
			Answered:   0,
			Total:      totalQuestions,
			Percentage: 0,
			Correct:    0,
			Score:      0,
		}
		api.WriteJSON(w, 200, response)
		return
	}

	// Calculate progress
	totalQuestions, err := qr.Repos.QuizQuestion.CountByQuizID(r.Context(), quizID)
	if err != nil {
		applog.Error("Failed to count questions:", err)
		api.WriteInternalError(w)
		return
	}

	answeredCount, err := qr.Repos.UserAnswer.CountByAttemptID(r.Context(), attempt.ID)
	if err != nil {
		applog.Error("Failed to count answers:", err)
		api.WriteInternalError(w)
		return
	}

	correctCount, err := qr.Repos.UserAnswer.CountCorrectByAttemptID(r.Context(), attempt.ID)
	if err != nil {
		applog.Error("Failed to count correct answers:", err)
		api.WriteInternalError(w)
		return
	}

	percentage := float64(answeredCount) / float64(totalQuestions) * 100
	score := float64(correctCount) / float64(totalQuestions) * 100

	response := QuizProgress{
		Answered:   answeredCount,
		Total:      totalQuestions,
		Percentage: percentage,
		Correct:    correctCount,
		Score:      score,
	}

	api.WriteJSON(w, 200, response)
}
