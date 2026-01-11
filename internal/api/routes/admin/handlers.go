package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/akramboussanni/gocode/internal/api"
	"github.com/akramboussanni/gocode/internal/api/routes/quiz"
	"github.com/akramboussanni/gocode/internal/applog"
	"github.com/akramboussanni/gocode/internal/model"
	quizpkg "github.com/akramboussanni/gocode/internal/quiz"
	"github.com/akramboussanni/gocode/internal/utils"
	"github.com/go-chi/chi/v5"
)

// @Summary List all users
// @Description Get a list of all users in the system (admin only)
// @Tags Admin
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {array} UserListResponse
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /admin/users [get]
// @Security BearerAuth
func (ar *AdminRouter) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Get all users
	users, err := ar.Repos.User.GetAllUsers(r.Context(), limit, offset)
	if err != nil {
		applog.Error("Failed to get users:", err)
		api.WriteInternalError(w)
		return
	}

	// Convert to response format
	response := make([]UserListResponse, 0, len(users))
	for _, user := range users {
		response = append(response, ToUserListResponse(user))
	}

	api.WriteJSON(w, 200, response)
}

// @Summary Get user details
// @Description Get detailed information about a specific user including progression (admin only)
// @Tags Admin
// @Produce json
// @Param userID path int true "User ID"
// @Success 200 {object} UserDetailResponse
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden"
// @Failure 404 {object} api.ErrorResponse "User not found"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /admin/users/{userID} [get]
// @Security BearerAuth
func (ar *AdminRouter) HandleGetUserDetail(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid user ID")
		return
	}

	// Get user
	user, err := ar.Repos.User.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		api.WriteNotFound(w, "User not found")
		return
	}

	// Get progression
	progression, err := ar.Repos.Progression.GetByUserID(r.Context(), userID)
	if err != nil {
		applog.Warn("Failed to get progression:", err)
		progression = &model.UserProgression{
			UserID:                userID,
			CurrentLevel:          1,
			TotalQuizzesCompleted: 0,
			TotalCoinsEarned:      0,
			StreakDays:            0,
		}
	}

	// Get coin balance
	coins, err := ar.Repos.Coin.GetUserCoins(r.Context(), userID)
	coinBalance := 0
	if err == nil && coins != nil {
		coinBalance = coins.Balance
	}

	response := UserDetailResponse{
		User:                  ToUserListResponse(user),
		CurrentLevel:          progression.CurrentLevel,
		TotalQuizzesCompleted: progression.TotalQuizzesCompleted,
		TotalCoinsEarned:      progression.TotalCoinsEarned,
		CoinBalance:           coinBalance,
		StreakDays:            progression.StreakDays,
		LastActivityDate:      progression.LastActivityDate,
	}

	api.WriteJSON(w, 200, response)
}

// @Summary Change user password
// @Description Change a user's password (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Param request body ChangePasswordRequest true "Password change request"
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse "Invalid request"
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden"
// @Failure 404 {object} api.ErrorResponse "User not found"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /admin/users/{userID}/password [post]
// @Security BearerAuth
func (ar *AdminRouter) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid user ID")
		return
	}

	// Decode request
	req, err := api.DecodeJSON[ChangePasswordRequest](w, r)
	if err != nil {
		return
	}

	// Validate password
	if len(req.NewPassword) < 8 {
		api.WriteBadRequest(w, "Password must be at least 8 characters")
		return
	}

	// Get user
	user, err := ar.Repos.User.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		api.WriteNotFound(w, "User not found")
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		applog.Error("Failed to hash password:", err)
		api.WriteInternalError(w)
		return
	}

	// Update password
	user.PasswordHash = hashedPassword
	user.JwtSessionID = time.Now().UnixMicro() // Invalidate all sessions

	if err := ar.Repos.User.UpdatePassword(r.Context(), user.ID, hashedPassword, user.JwtSessionID); err != nil {
		applog.Error("Failed to update password:", err)
		api.WriteInternalError(w)
		return
	}

	api.WriteMessage(w, 200, "message", "Password changed successfully")
}

// @Summary Delete user
// @Description Delete a user account (admin only)
// @Tags Admin
// @Produce json
// @Param userID path int true "User ID"
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse "Invalid request"
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden"
// @Failure 404 {object} api.ErrorResponse "User not found"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /admin/users/{userID} [delete]
// @Security BearerAuth
func (ar *AdminRouter) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid user ID")
		return
	}

	// Prevent deleting yourself
	admin, _ := utils.UserFromContext(r.Context())
	if admin.ID == userID {
		api.WriteBadRequest(w, "Cannot delete your own account")
		return
	}

	// Check user exists
	user, err := ar.Repos.User.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		api.WriteNotFound(w, "User not found")
		return
	}

	// Delete user
	if err := ar.Repos.User.DeleteUser(r.Context(), userID); err != nil {
		applog.Error("Failed to delete user:", err)
		api.WriteInternalError(w)
		return
	}

	api.WriteMessage(w, 200, "message", "User deleted successfully")
}

// @Summary Create quiz
// @Description Create a new quiz (admin can set all fields)
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body CreateQuizRequest true "Quiz creation request"
// @Success 201 {object} model.Quiz
// @Failure 400 {object} api.ErrorResponse "Invalid request"
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /admin/quizzes [post]
// @Security BearerAuth
func (ar *AdminRouter) HandleCreateQuiz(w http.ResponseWriter, r *http.Request) {
	admin, _ := utils.UserFromContext(r.Context())

	// Decode request
	req, err := api.DecodeJSON[CreateQuizRequest](w, r)
	if err != nil {
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
				deck, err := ar.DeckService.GetDeck(r.Context(), sel.DeckID)
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

		autoQuiz, err := ar.QuizService.CreateQuiz(r.Context(), config, parsedDecks)
		if err != nil {
			applog.Error("Failed to auto-generate quiz questions:", err)
			api.WriteInternalError(w)
			return
		}

		// Add auto-generated questions to the list
		allQuestions = append(allQuestions, autoQuiz.Questions...)
	}

	// Generate quiz ID
	quizID := utils.GenerateSnowflakeID()

	// Create quiz record
	quizModel := &model.Quiz{
		ID:                 quizID,
		Title:              req.Title,
		Description:        req.Description,
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
		CreatedBy:          &admin.ID,
		CreatedAt:          time.Now().Unix(),
		IsActive:           true,
	}

	// Begin transaction
	tx, err := ar.Repos.Quiz.BeginTx(r.Context())
	if err != nil {
		applog.Error("Failed to begin transaction:", err)
		api.WriteInternalError(w)
		return
	}
	defer tx.Rollback()

	// Create quiz
	if err := ar.Repos.Quiz.CreateWithTx(r.Context(), tx, quizModel); err != nil {
		applog.Error("Failed to create quiz:", err)
		api.WriteInternalError(w)
		return
	}

	// Store deck selections if auto-generation was used
	if req.AutoGenerate != nil {
		for _, sel := range req.AutoGenerate.DeckSelections {
			// Get categories for this deck
			categories, err := ar.Repos.Category.GetByDeckID(r.Context(), sel.DeckID)
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
				if err := ar.Repos.QuizCategorySelection.CreateWithTx(r.Context(), tx, selection); err != nil {
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

		if err := ar.Repos.QuizQuestion.CreateWithTx(r.Context(), tx, quizQuestion); err != nil {
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
	var questions []quiz.QuizQuestionResponse
	for _, q := range allQuestions {
		questions = append(questions, quiz.QuizQuestionResponse{
			ID:           q.ID,
			QuestionText: q.QuestionText,
			Options:      q.Options,
			QuestionType: string(q.QuestionType),
			Direction:    string(q.Direction),
		})
	}

	response := quiz.QuizDetailResponse{
		ID:           quizID,
		Title:        req.Title,
		Description:  req.Description,
		Config:       quizpkg.QuizConfig{}, // Empty config for response
		Questions:    questions,
		CurrentIndex: 0,
		StartedAt:    nil,
		Progress: quiz.QuizProgress{
			Answered:   0,
			Total:      len(allQuestions),
			Percentage: 0,
		},
	}

	api.WriteJSON(w, 201, response)
}

// @Summary Get quiz statistics
// @Description Get statistics for all quizzes (admin only)
// @Tags Admin
// @Produce json
// @Param limit query int false "Items per page" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} QuizStatsResponse
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /admin/quizzes/stats [get]
// @Security BearerAuth
func (ar *AdminRouter) HandleGetQuizStats(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	// Get all quizzes
	quizzes, err := ar.Repos.Quiz.GetAllQuizzes(r.Context(), limit, offset)
	if err != nil {
		applog.Error("Failed to get quizzes:", err)
		api.WriteInternalError(w)
		return
	}

	response := make([]QuizStatsResponse, 0, len(quizzes))
	for _, quiz := range quizzes {
		stats := QuizStatsResponse{
			QuizID:    quiz.ID,
			Title:     quiz.Title,
			CreatedBy: quiz.CreatedBy,
			IsSystem:  quiz.IsSystem,
			CreatedAt: quiz.CreatedAt,
		}

		// Get creator username if available
		if quiz.CreatedBy != nil {
			creator, err := ar.Repos.User.GetUserByID(r.Context(), *quiz.CreatedBy)
			if err == nil && creator != nil {
				stats.CreatorUsername = &creator.Username
			}
		}

		// Get attempt stats
		totalAttempts, _ := ar.Repos.QuizAttempt.CountAttemptsByQuiz(r.Context(), quiz.ID)
		stats.TotalAttempts = totalAttempts

		if totalAttempts > 0 {
			avgScore, _ := ar.Repos.QuizAttempt.GetAverageScoreByQuiz(r.Context(), quiz.ID)
			if avgScore != nil {
				stats.AverageScore = *avgScore
			}

			passCount, _ := ar.Repos.QuizAttempt.CountPassedAttempts(r.Context(), quiz.ID)
			stats.PassRate = (float64(passCount) / float64(totalAttempts)) * 100
		}

		response = append(response, stats)
	}

	api.WriteJSON(w, 200, response)
}

// @Summary List user-generated quizzes
// @Description Get all user-generated quizzes (admin only)
// @Tags Admin
// @Produce json
// @Param limit query int false "Items per page" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} UserQuizResponse
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /admin/quizzes/user-generated [get]
// @Security BearerAuth
func (ar *AdminRouter) HandleListUserQuizzes(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	// Get user-generated quizzes
	quizzes, err := ar.Repos.Quiz.GetUserGeneratedQuizzes(r.Context(), limit, offset)
	if err != nil {
		applog.Error("Failed to get user quizzes:", err)
		api.WriteInternalError(w)
		return
	}

	response := make([]UserQuizResponse, 0, len(quizzes))
	for _, quiz := range quizzes {
		resp := UserQuizResponse{
			ID:          quiz.ID,
			Title:       quiz.Title,
			Description: quiz.Description,
			IsPublic:    quiz.IsPublic,
			CreatedAt:   quiz.CreatedAt,
		}

		if quiz.CreatedBy != nil {
			resp.CreatedBy = *quiz.CreatedBy
			// Get creator username
			creator, err := ar.Repos.User.GetUserByID(r.Context(), *quiz.CreatedBy)
			if err == nil && creator != nil {
				resp.CreatorUsername = creator.Username
			}
		}

		// Get attempt count
		count, _ := ar.Repos.QuizAttempt.CountAttemptsByQuiz(r.Context(), quiz.ID)
		resp.TotalAttempts = count

		response = append(response, resp)
	}

	api.WriteJSON(w, 200, response)
}

// @Summary Delete quiz
// @Description Delete a user-generated quiz (admin only)
// @Tags Admin
// @Produce json
// @Param quizID path int true "Quiz ID"
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse "Invalid request"
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden - Cannot delete system quizzes"
// @Failure 404 {object} api.ErrorResponse "Quiz not found"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /admin/quizzes/{quizID} [delete]
// @Security BearerAuth
func (ar *AdminRouter) HandleDeleteQuiz(w http.ResponseWriter, r *http.Request) {
	quizID, err := strconv.ParseInt(chi.URLParam(r, "quizID"), 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid quiz ID")
		return
	}

	// Get quiz
	quiz, err := ar.Repos.Quiz.GetByID(r.Context(), quizID)
	if err != nil || quiz == nil {
		api.WriteNotFound(w, "Quiz not found")
		return
	}

	// Prevent deleting system quizzes
	if quiz.IsSystem {
		api.WriteMessage(w, 403, "error", "Cannot delete system quizzes")
		return
	}

	// Soft delete by setting is_active = false
	if err := ar.Repos.Quiz.DeactivateQuiz(r.Context(), quizID); err != nil {
		applog.Error("Failed to delete quiz:", err)
		api.WriteInternalError(w)
		return
	}

	api.WriteMessage(w, 200, "message", "Quiz deleted successfully")
}

// @Summary Get system statistics
// @Description Get overall system statistics (admin only)
// @Tags Admin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /admin/stats/overview [get]
// @Security BearerAuth
func (ar *AdminRouter) HandleGetSystemStats(w http.ResponseWriter, r *http.Request) {
	stats := make(map[string]interface{})

	// Total users
	totalUsers, err := ar.Repos.User.CountTotalUsers(r.Context())
	if err == nil {
		stats["total_users"] = totalUsers
	}

	// Total quizzes
	totalQuizzes, err := ar.Repos.Quiz.CountTotalQuizzes(r.Context())
	if err == nil {
		stats["total_quizzes"] = totalQuizzes
	}

	// Total attempts
	totalAttempts, err := ar.Repos.QuizAttempt.CountTotalAttempts(r.Context())
	if err == nil {
		stats["total_attempts"] = totalAttempts
	}

	// Active users (last 7 days)
	activeUsers, err := ar.Repos.User.CountActiveUsers(r.Context(), 7)
	if err == nil {
		stats["active_users_7d"] = activeUsers
	}

	api.WriteJSON(w, 200, stats)
}
