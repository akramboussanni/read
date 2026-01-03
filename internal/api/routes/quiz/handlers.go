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
	api.WriteMessage(w, 501, "error", "Not implemented yet")
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
	_ = user
	api.WriteMessage(w, 501, "error", "Not implemented yet")
}

func (qr *QuizRouter) HandleCreateQuiz(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	userID := user.ID

	// Require verified email for quiz creation
	if !user.EmailConfirmed {
		api.WriteMessage(w, 403, "error", "Email verification required to create quizzes. Please add and verify your email.")
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
	api.WriteMessage(w, 501, "error", "Not implemented yet")
}

func (qr *QuizRouter) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	user, ok := utils.UserFromContext(r.Context())
	if !ok {
		api.WriteUnauthorized(w)
		return
	}
	_ = user
	api.WriteMessage(w, 501, "error", "Not implemented yet")
}
