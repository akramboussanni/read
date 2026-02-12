package admin

import (
	"time"

	"github.com/akramboussanni/gocode/internal/middleware"
	quiz "github.com/akramboussanni/gocode/internal/quiz"
	"github.com/akramboussanni/gocode/internal/repo"
	"github.com/go-chi/chi/v5"
)

type AdminRouter struct {
	Repos       *repo.Repos
	QuizService *quiz.QuizService
	DeckService *quiz.DeckService
}

func NewAdminRouter(repos *repo.Repos, quizService *quiz.QuizService, deckService *quiz.DeckService) *AdminRouter {
	return &AdminRouter{
		Repos:       repos,
		QuizService: quizService,
		DeckService: deckService,
	}
}

func (ar *AdminRouter) Routes() chi.Router {
	r := chi.NewRouter()

	// Apply admin-only middleware
	middleware.RequireAdmin(r)
	middleware.AddRatelimit(r, 30, 1*time.Minute) // 30 requests per minute for admins

	// User management
	r.Get("/users", ar.HandleListUsers)
	r.Get("/users/{userID}", ar.HandleGetUserDetail)
	r.Post("/users/{userID}/password", ar.HandleChangePassword)
	r.Delete("/users/{userID}", ar.HandleDeleteUser)

	// Quiz management
	r.Post("/quizzes", ar.HandleCreateQuiz)
	r.Get("/quizzes/stats", ar.HandleGetQuizStats)
	r.Get("/quizzes/user-generated", ar.HandleListUserQuizzes)
	r.Put("/quizzes/{quizID}", ar.HandleUpdateQuiz)
	r.Delete("/quizzes/{quizID}", ar.HandleDeleteQuiz)

	// System statistics
	r.Get("/stats/overview", ar.HandleGetSystemStats)

	return r
}
