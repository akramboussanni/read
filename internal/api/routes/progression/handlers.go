package progression

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/akramboussanni/gocode/internal/api"
	"github.com/akramboussanni/gocode/internal/applog"
	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/utils"
	"github.com/go-chi/chi/v5"
)

// @Summary Get user's progression status
// @Description Retrieves the current user's progression status including level, stats, and next quiz
// @Tags Progression
// @Produce json
// @Success 200 {object} ProgressionStatusResponse
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /progression/status [get]
// @Security BearerAuth
func (pr *ProgressionRouter) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	userID := user.ID

	// Get user progression
	progression, err := pr.Repos.Progression.GetByUserID(r.Context(), userID)
	if err != nil {
		applog.Error("Failed to get progression:", err)
		api.WriteInternalError(w)
		return
	}

	// Get coin balance
	coins, err := pr.Repos.Coin.GetUserCoins(r.Context(), userID)
	if err != nil {
		applog.Error("Failed to get coins:", err)
		// Continue anyway, default to 0
		coins = nil
	}

	coinBalance := 0
	if coins != nil {
		coinBalance = coins.Balance
	}

	// Get next available quiz
	nextQuiz, err := pr.Repos.Quiz.GetNextUnlockedQuiz(r.Context(), userID, progression.CurrentLevel)
	if err != nil {
		applog.Warn("Failed to get next quiz:", err)
	}

	response := ProgressionStatusResponse{
		CurrentLevel:          progression.CurrentLevel,
		TotalQuizzesCompleted: progression.TotalQuizzesCompleted,
		TotalCoinsEarned:      progression.TotalCoinsEarned,
		StreakDays:            progression.StreakDays,
		CoinBalance:           coinBalance,
	}

	if nextQuiz != nil {
		response.NextQuiz = &QuizPreview{
			ID:          nextQuiz.ID,
			Title:       nextQuiz.Title,
			Description: nextQuiz.Description,
			Level:       nextQuiz.LevelOrder,
			CoinReward:  nextQuiz.CoinReward,
		}
	}

	api.WriteJSON(w, 200, response)
}

// @Summary List progression quizzes
// @Description Lists all system quizzes with lock status and completion info
// @Tags Progression
// @Produce json
// @Success 200 {array} QuizPreview
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /progression/quizzes [get]
// @Security BearerAuth
func (pr *ProgressionRouter) HandleListQuizzes(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	userID := user.ID

	// Get user's unlocked quiz IDs
	progression, err := pr.Repos.Progression.GetByUserID(r.Context(), userID)
	if err != nil {
		applog.Error("Failed to get progression:", err)
		api.WriteInternalError(w)
		return
	}

	var unlockedIDs []int64
	if progression.UnlockedQuizIDs != "" {
		if err := json.Unmarshal([]byte(progression.UnlockedQuizIDs), &unlockedIDs); err != nil {
			applog.Warn("Failed to parse unlocked quiz IDs:", err)
			unlockedIDs = []int64{}
		}
	}

	// Get all system quizzes
	quizzes, err := pr.Repos.Quiz.GetSystemQuizzes(r.Context())
	if err != nil {
		applog.Error("Failed to get quizzes:", err)
		api.WriteInternalError(w)
		return
	}

	// Get user's best attempts
	attempts, err := pr.Repos.QuizAttempt.GetBestAttemptsByUser(r.Context(), userID)
	if err != nil {
		applog.Warn("Failed to get attempts:", err)
		attempts = make(map[int64]float64)
	}

	// Build response
	previews := make([]QuizPreview, 0, len(quizzes))
	for _, quiz := range quizzes {
		isUnlocked := contains(unlockedIDs, quiz.ID) || quiz.LevelOrder == 0

		preview := QuizPreview{
			ID:          quiz.ID,
			Title:       quiz.Title,
			Description: quiz.Description,
			Level:       quiz.LevelOrder,
			IsLocked:    !isUnlocked,
			CoinReward:  quiz.CoinReward,
		}

		if score, exists := attempts[quiz.ID]; exists {
			preview.IsCompleted = true
			preview.BestScore = &score
			percentage := (score / float64(preview.QuestionCount)) * 100
			preview.BestPercentage = &percentage
		}

		previews = append(previews, preview)
	}

	api.WriteJSON(w, 200, previews)
}

// @Summary Get specific quiz details
// @Description Get details of a specific system quiz if unlocked
// @Tags Progression
// @Produce json
// @Param quizID path int true "Quiz ID"
// @Success 200 {object} QuizPreview
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Quiz is locked"
// @Failure 404 {object} api.ErrorResponse "Quiz not found"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /progression/quiz/{quizID} [get]
// @Security BearerAuth
func (pr *ProgressionRouter) HandleGetQuiz(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	userID := user.ID

	quizID, err := strconv.ParseInt(chi.URLParam(r, "quizID"), 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid quiz ID")
		return
	}

	// Check if quiz is unlocked
	isUnlocked, err := pr.Repos.Progression.IsQuizUnlocked(r.Context(), userID, quizID)
	if err != nil {
		applog.Error("Failed to check quiz unlock status:", err)
		api.WriteInternalError(w)
		return
	}

	if !isUnlocked {
		api.WriteMessage(w, 403, "error", "Quiz is locked. Complete previous quizzes first.")
		return
	}

	// Get quiz details
	quiz, err := pr.Repos.Quiz.GetByID(r.Context(), quizID)
	if err != nil || quiz == nil {
		api.WriteNotFound(w, "Quiz not found")
		return
	}

	// Get question count
	count, err := pr.Repos.Question.CountByQuizID(r.Context(), quizID)
	if err != nil {
		applog.Warn("Failed to count questions:", err)
		count = 0
	}

	// Get user's best attempt
	attempts, err := pr.Repos.QuizAttempt.GetBestAttemptsByUser(r.Context(), userID)
	if err != nil {
		applog.Warn("Failed to get best attempts:", err)
		attempts = make(map[int64]float64)
	}

	preview := QuizPreview{
		ID:            quiz.ID,
		Title:         quiz.Title,
		Description:   quiz.Description,
		Level:         quiz.LevelOrder,
		IsLocked:      false,
		CoinReward:    quiz.CoinReward,
		QuestionCount: count,
	}

	if score, exists := attempts[quizID]; exists {
		preview.IsCompleted = true
		preview.BestScore = &score
		// Get actual max score by summing question points
		questions, err := pr.Repos.Question.GetQuestionsByQuizID(r.Context(), quizID)
		if err == nil && len(questions) > 0 {
			var maxScore float64
			for _, q := range questions {
				maxScore += float64(q.Points)
			}
			if maxScore > 0 {
				percentage := (score / maxScore) * 100
				preview.BestPercentage = &percentage
			}
		}
	}

	api.WriteJSON(w, 200, preview)
}

// @Summary Start a quiz
// @Description Start a new quiz attempt
// @Tags Progression
// @Accept json
// @Produce json
// @Param quizID path int true "Quiz ID"
// @Param request body StartQuizRequest true "Start quiz request"
// @Success 200 {object} StartQuizResponse
// @Failure 400 {object} api.ErrorResponse "Invalid request"
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Quiz is locked"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /progression/quiz/{quizID}/start [post]
// @Security BearerAuth
func (pr *ProgressionRouter) HandleStartQuiz(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	userID := user.ID

	quizID, err := strconv.ParseInt(chi.URLParam(r, "quizID"), 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid quiz ID")
		return
	}

	// Check if quiz is unlocked
	isUnlocked, err := pr.Repos.Progression.IsQuizUnlocked(r.Context(), userID, quizID)
	if err != nil {
		applog.Error("Failed to check quiz unlock status:", err)
		api.WriteInternalError(w)
		return
	}

	if !isUnlocked {
		api.WriteMessage(w, 403, "error", "Quiz is locked. Complete previous quizzes first.")
		return
	}

	// Get quiz details
	quiz, err := pr.Repos.Quiz.GetByID(r.Context(), quizID)
	if err != nil || quiz == nil {
		api.WriteNotFound(w, "Quiz not found")
		return
	}

	// Create quiz attempt
	attemptID := utils.GenerateSnowflakeID()
	attempt := &model.QuizAttempt{
		ID:        attemptID,
		UserID:    userID,
		QuizID:    quizID,
		StartedAt: time.Now().Unix(),
	}

	if err := pr.Repos.QuizAttempt.Create(r.Context(), attempt); err != nil {
		applog.Error("Failed to create quiz attempt:", err)
		api.WriteInternalError(w)
		return
	}

	// Get questions for the quiz
	questions, err := pr.Repos.Question.GetQuestionsByQuizID(r.Context(), quizID)
	if err != nil {
		applog.Error("Failed to get questions:", err)
		api.WriteInternalError(w)
		return
	}

	// Build response with questions and options
	questionResponses := make([]QuestionWithOptions, 0, len(questions))
	for _, q := range questions {
		qResp := QuestionWithOptions{
			ID:           q.ID,
			QuestionText: q.QuestionText,
			QuestionType: q.QuestionType,
			Points:       q.Points,
		}

		// For MCQ questions, get options
		if q.QuestionType == "mcq" {
			options, err := pr.Repos.QuestionOption.GetByQuestionID(r.Context(), q.ID)
			if err != nil {
				applog.Warn("Failed to get options for question:", q.ID, err)
			} else {
				// Convert to response format (don't reveal correct answer)
				for _, opt := range options {
					qResp.Options = append(qResp.Options, Option{
						ID:         opt.ID,
						OptionText: opt.OptionText,
					})
				}
			}
		}

		questionResponses = append(questionResponses, qResp)
	}

	response := StartQuizResponse{
		AttemptID: attemptID,
		QuizID:    quizID,
		Title:     quiz.Title,
		TimeLimit: quiz.TimeLimit,
		Questions: questionResponses,
	}

	api.WriteJSON(w, 200, response)
}

// @Summary Submit quiz answers
// @Description Submit answers for a quiz attempt
// @Tags Progression
// @Accept json
// @Produce json
// @Param quizID path int true "Quiz ID"
// @Param request body SubmitQuizRequest true "Quiz answers"
// @Success 200 {object} SubmitQuizResponse
// @Failure 400 {object} api.ErrorResponse "Invalid request"
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 500 {object} api.ErrorResponse "Internal server error"
// @Router /progression/quiz/{quizID}/submit [post]
// @Security BearerAuth
func (pr *ProgressionRouter) HandleSubmitQuiz(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	userID := user.ID

	quizID, err := strconv.ParseInt(chi.URLParam(r, "quizID"), 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid quiz ID")
		return
	}

	// Decode request
	req, err := api.DecodeJSON[SubmitQuizRequest](w, r)
	if err != nil {
		return
	}

	// Validate attempt exists and belongs to user
	attempt, err := pr.Repos.QuizAttempt.GetByID(r.Context(), req.AttemptID)
	if err != nil || attempt == nil {
		api.WriteNotFound(w, "Quiz attempt not found")
		return
	}

	if attempt.UserID != userID {
		api.WriteUnauthorized(w)
		return
	}

	if attempt.CompletedAt != nil {
		api.WriteBadRequest(w, "Quiz attempt already completed")
		return
	}

	// Get quiz details
	quiz, err := pr.Repos.Quiz.GetByID(r.Context(), quizID)
	if err != nil || quiz == nil {
		api.WriteNotFound(w, "Quiz not found")
		return
	}

	// Get all questions for the quiz
	questions, err := pr.Repos.Question.GetQuestionsByQuizID(r.Context(), quizID)
	if err != nil {
		applog.Error("Failed to get questions:", err)
		api.WriteInternalError(w)
		return
	}

	// Create question map for quick lookup
	questionMap := make(map[int64]*model.Question)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	// Grade answers
	var totalScore float64
	var maxScore float64
	userAnswers := make([]*model.UserAnswer, 0, len(req.Answers))
	results := make([]AnswerResult, 0, len(req.Answers))

	for _, ans := range req.Answers {
		question, exists := questionMap[ans.QuestionID]
		if !exists {
			continue // Skip invalid questions
		}

		maxScore += float64(question.Points)

		var isCorrect bool
		var pointsEarned float64
		var correctAnswer string

		if question.QuestionType == "mcq" {
			// For MCQ, check if selected option is correct
			optionID, _ := strconv.ParseInt(ans.Answer, 10, 64)
			options, err := pr.Repos.QuestionOption.GetByQuestionID(r.Context(), question.ID)
			if err == nil {
				for _, opt := range options {
					if opt.ID == optionID && opt.IsCorrect {
						isCorrect = true
						pointsEarned = float64(question.Points)
					}
					if opt.IsCorrect {
						correctAnswer = opt.OptionText
					}
				}
			}
		} else {
			// For written answers, do case-insensitive match with trimmed whitespace
			correctAnswer = question.CorrectAnswer
			userAnswerNormalized := strings.ToLower(strings.TrimSpace(ans.Answer))
			correctAnswerNormalized := strings.ToLower(strings.TrimSpace(question.CorrectAnswer))
			isCorrect = userAnswerNormalized == correctAnswerNormalized
			if isCorrect {
				pointsEarned = float64(question.Points)
			}
		}

		if isCorrect {
			totalScore += pointsEarned
		}

		// Store user answer
		userAnswer := &model.UserAnswer{
			ID:           utils.GenerateSnowflakeID(),
			AttemptID:    req.AttemptID,
			QuestionID:   ans.QuestionID,
			UserAnswer:   ans.Answer,
			IsCorrect:    isCorrect,
			PointsEarned: pointsEarned,
			AnsweredAt:   time.Now().Unix(),
		}
		userAnswers = append(userAnswers, userAnswer)

		// Build result
		results = append(results, AnswerResult{
			QuestionID:    ans.QuestionID,
			UserAnswer:    ans.Answer,
			CorrectAnswer: correctAnswer,
			IsCorrect:     isCorrect,
			PointsEarned:  pointsEarned,
		})
	}

	// Calculate percentage and pass/fail
	var percentage float64
	if maxScore > 0 {
		percentage = (totalScore / maxScore) * 100
	}

	passPercentage := 70 // Default
	if quiz.PassPercentage != nil {
		passPercentage = *quiz.PassPercentage
	}
	passed := percentage >= float64(passPercentage)

	// Calculate coins earned
	var coinsEarned int
	if passed && quiz.GivesCoins {
		// Count correct answers
		correctCount := 0
		for _, result := range results {
			if result.IsCorrect {
				correctCount++
			}
		}
		coinsEarned = correctCount * quiz.CoinReward
	}

	// Update attempt
	completedAt := time.Now().Unix()
	timeTaken := int(completedAt - attempt.StartedAt)
	attempt.CompletedAt = &completedAt
	attempt.Score = &totalScore
	attempt.MaxScore = &maxScore
	attempt.Percentage = &percentage
	attempt.Passed = &passed
	attempt.TimeTaken = &timeTaken
	attempt.CoinsEarned = coinsEarned

	if err := pr.Repos.QuizAttempt.Update(r.Context(), attempt); err != nil {
		applog.Error("Failed to update attempt:", err)
		api.WriteInternalError(w)
		return
	}

	// Save user answers
	if err := pr.Repos.UserAnswer.BatchCreate(r.Context(), userAnswers); err != nil {
		applog.Error("Failed to save user answers:", err)
		// Continue anyway, answers are secondary
	}

	// Award coins if applicable
	if coinsEarned > 0 {
		coins, err := pr.Repos.Coin.GetUserCoins(r.Context(), userID)
		if err != nil || coins == nil {
			coins = &model.UserCoins{
				UserID:         userID,
				Balance:        0,
				LifetimeEarned: 0,
			}
		}

		coins.Balance += coinsEarned
		coins.LifetimeEarned += coinsEarned
		coins.LastUpdated = time.Now().Unix()

		if err := pr.Repos.Coin.CreateOrUpdateCoins(r.Context(), coins); err != nil {
			applog.Error("Failed to update coins:", err)
		}

		// Record transaction
		refType := "quiz_attempt"
		desc := "Completed quiz: " + quiz.Title
		tx := &model.CoinTransaction{
			ID:              utils.GenerateSnowflakeID(),
			UserID:          userID,
			Amount:          coinsEarned,
			TransactionType: "earn",
			ReferenceType:   &refType,
			ReferenceID:     &req.AttemptID,
			Description:     &desc,
			CreatedAt:       time.Now().Unix(),
		}
		if err := pr.Repos.Coin.AddTransaction(r.Context(), tx); err != nil {
			applog.Error("Failed to record coin transaction:", err)
		}
	}

	// Update progression if quiz passed and is a system quiz
	var nextUnlocked *int64
	if passed && quiz.IsSystem {
		if err := pr.Repos.Progression.CompleteQuiz(r.Context(), userID, quizID, coinsEarned); err != nil {
			applog.Error("Failed to update progression:", err)
		} else {
			// Check if next quiz was unlocked
			progression, _ := pr.Repos.Progression.GetByUserID(r.Context(), userID)
			if progression != nil {
				nextQuiz, _ := pr.Repos.Quiz.GetNextUnlockedQuiz(r.Context(), userID, progression.CurrentLevel)
				if nextQuiz != nil {
					nextUnlocked = &nextQuiz.ID
				}
			}
		}
	}

	response := SubmitQuizResponse{
		Score:        totalScore,
		MaxScore:     maxScore,
		Percentage:   percentage,
		Passed:       passed,
		CoinsEarned:  coinsEarned,
		NextUnlocked: nextUnlocked,
		Results:      results,
	}

	api.WriteJSON(w, 200, response)
}

// Helper function
func contains(slice []int64, item int64) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
