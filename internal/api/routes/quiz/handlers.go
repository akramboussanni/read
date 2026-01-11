package quiz

import (
	"net/http"
	"strconv"

	"github.com/akramboussanni/gocode/internal/api"
	"github.com/akramboussanni/gocode/internal/applog"
	quizpkg "github.com/akramboussanni/gocode/internal/quiz"
	"github.com/akramboussanni/gocode/internal/utils"
	"github.com/go-chi/chi/v5"
)

func (qr *QuizRouter) HandleListQuizzes(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	quizzes, err := qr.Repos.Quiz.GetAllQuizzes(r.Context(), limit, offset)
	if err != nil {
		applog.Error("Failed to get quizzes:", err)
		api.WriteInternalError(w)
		return
	}

	response := []QuizListItem{}
	for _, quiz := range quizzes {
		questionCount, _ := qr.Repos.Question.CountByQuizID(r.Context(), quiz.ID)

		var creatorName string
		if quiz.CreatedBy != nil {
			creator, err := qr.Repos.User.GetUserByID(r.Context(), *quiz.CreatedBy)
			if err == nil {
				creatorName = creator.Username
			}
		}

		response = append(response, QuizListItem{
			ID:            quiz.ID,
			Title:         quiz.Title,
			Description:   quiz.Description,
			CreatorName:   creatorName,
			QuestionCount: questionCount,
			IsPublic:      quiz.IsPublic,
			CreatedAt:     quiz.CreatedAt,
		})
	}

	api.WriteJSON(w, 200, response)
}

func (qr *QuizRouter) HandleGetQuiz(w http.ResponseWriter, r *http.Request) {
	quizID, err := strconv.ParseInt(chi.URLParam(r, "quizID"), 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid quiz ID")
		return
	}
	_ = quizID
	api.WriteMessage(w, 501, "error", "Not implemented yet")
}

func (qr *QuizRouter) HandleGetMyQuizzes(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}

	quizzes, err := qr.Repos.Quiz.GetQuizzesByCreator(r.Context(), user.ID)
	if err != nil {
		applog.Error("Failed to get user quizzes:", err)
		api.WriteInternalError(w)
		return
	}

	response := []QuizListItem{}
	for _, quiz := range quizzes {
		questionCount, _ := qr.Repos.Question.CountByQuizID(r.Context(), quiz.ID)

		response = append(response, QuizListItem{
			ID:            quiz.ID,
			Title:         quiz.Title,
			Description:   quiz.Description,
			QuestionCount: questionCount,
			IsPublic:      quiz.IsPublic,
			CreatedAt:     quiz.CreatedAt,
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
	userID := user.ID

	// Require email to exist and be verified for quiz creation
	if user.Email == "" {
		api.WriteMessage(w, 403, "error", "Email address required to create quizzes. Please add an email to your account.")
		return
	}

	if !user.EmailConfirmed {
		api.WriteMessage(w, 403, "error", "Email verification required to create quizzes. Please verify your email address.")
		return
	}

	req, err := api.DecodeJSON[CreateQuizRequest](w, r)
	if err != nil {
		applog.Error("Failed to decode quiz request:", err)
		return
	}

	if req.Title == "" {
		api.WriteBadRequest(w, "Title is required")
		return
	}

	serviceReq := quizpkg.CreateQuizRequest{
		Title:            req.Title,
		Description:      req.Description,
		DeckID:           req.DeckID,
		QuestionMode:     req.QuestionMode,
		TimeLimit:        req.TimeLimit,
		PassPercentage:   req.PassPercentage,
		ShuffleQuestions: req.ShuffleQuestions,
		IsSystem:         false,
		CreatedBy:        &userID,
	}

	for _, cs := range req.CategorySelections {
		serviceReq.CategorySelections = append(serviceReq.CategorySelections, quizpkg.CategorySelection{
			CategoryID:    cs.CategoryID,
			QuestionCount: cs.QuestionCount,
		})
	}

	isAdmin := user.Role == "admin"
	quiz, err := qr.QuizService.CreateQuiz(r.Context(), serviceReq, isAdmin)
	if err != nil {
		applog.Error("Failed to create quiz:", err)
		api.WriteInternalError(w)
		return
	}

	response := QuizDetailResponse{
		ID:               quiz.ID,
		Title:            quiz.Title,
		Description:      quiz.Description,
		CreatorID:        quiz.CreatedBy,
		ShuffleQuestions: quiz.ShuffleQuestions,
		IsPublic:         req.IsPublic,
		IsSystem:         quiz.IsSystem,
		GivesCoins:       quiz.GivesCoins,
		CoinReward:       quiz.CoinReward,
		CreatedAt:        quiz.CreatedAt,
	}

	api.WriteJSON(w, 201, response)
}

func (qr *QuizRouter) HandleUpdateQuiz(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	_ = user
	api.WriteMessage(w, 501, "error", "Not implemented yet")
}

func (qr *QuizRouter) HandleDeleteQuiz(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	_ = user
	api.WriteMessage(w, 501, "error", "Not implemented yet")
}

func (qr *QuizRouter) HandleStartQuiz(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	_ = user
	api.WriteMessage(w, 501, "error", "Not implemented yet")
}

func (qr *QuizRouter) HandleSubmitQuiz(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	_ = user
	api.WriteMessage(w, 501, "error", "Not implemented yet")
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

	serviceReq := quizpkg.CreateQuestionRequest{
		DeckID:           req.DeckID,
		CategoryID:       req.CategoryID,
		QuestionKey:      req.QuestionKey,
		QuestionText:     req.QuestionText,
		CorrectAnswer:    req.CorrectAnswer,
		QuestionType:     req.QuestionType,
		AnswerMode:       req.AnswerMode,
		ManualAnswers:    req.ManualAnswers,
		GenerationRule:   req.GenerationRule,
		DataSourceName:   req.DataSourceName,
		WrongAnswerCount: req.WrongAnswerCount,
	}

	question, err := qr.QuizService.CreateQuestion(r.Context(), serviceReq)
	if err != nil {
		applog.Error("Failed to create question:", err)
		api.WriteInternalError(w)
		return
	}

	api.WriteJSON(w, 201, question)
}

func (qr *QuizRouter) HandleListDecks(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	_ = user

	decks, err := qr.Repos.Deck.GetAll(r.Context())
	if err != nil {
		applog.Error("Failed to get decks:", err)
		api.WriteInternalError(w)
		return
	}

	response := []DeckListItem{}
	for _, deck := range decks {
		categoryCount, _ := qr.Repos.Category.CountByDeckID(r.Context(), deck.ID)
		questionCount, _ := qr.Repos.Question.CountByDeckID(r.Context(), deck.ID)

		response = append(response, DeckListItem{
			ID:            deck.ID,
			DeckKey:       deck.DeckKey,
			Title:         deck.Title,
			CategoryCount: categoryCount,
			QuestionCount: questionCount,
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

	categories, err := qr.Repos.Category.GetByDeckID(r.Context(), deckID)
	if err != nil {
		applog.Error("Failed to get categories:", err)
		api.WriteInternalError(w)
		return
	}

	response := []CategoryListItem{}
	for _, cat := range categories {
		questionCount, _ := qr.Repos.Question.CountByCategoryID(r.Context(), cat.ID)

		response = append(response, CategoryListItem{
			ID:            cat.ID,
			CategoryKey:   cat.CategoryKey,
			Title:         cat.Title,
			QuestionCount: questionCount,
		})
	}

	api.WriteJSON(w, 200, response)
}
