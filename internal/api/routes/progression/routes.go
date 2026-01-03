package progression

import (
	"net/http"
	"time"

	"github.com/akramboussanni/gocode/internal/middleware"
	"github.com/akramboussanni/gocode/internal/quiz"
	"github.com/akramboussanni/gocode/internal/repo"
	"github.com/go-chi/chi/v5"
)

type ProgressionRouter struct {
	UserRepo    *repo.UserRepo
	TokenRepo   *repo.TokenRepo
	QuizService *quiz.QuizService
	Repos       *repo.Repos
}

func NewProgressionRouter(repos *repo.Repos, quizService *quiz.QuizService) http.Handler {
	pr := &ProgressionRouter{
		UserRepo:    repos.User,
		TokenRepo:   repos.Token,
		QuizService: quizService,
		Repos:       repos,
	}
	r := chi.NewRouter()

	r.Use(middleware.MaxBytesMiddleware(1 << 20))

	// All progression routes require authentication
	r.Group(func(r chi.Router) {
		middleware.AddAuth(r, pr.UserRepo, pr.TokenRepo)
		middleware.AddRatelimit(r, 30, 1*time.Minute)

		r.Get("/status", pr.HandleGetStatus)
		r.Get("/quizzes", pr.HandleListQuizzes)
		r.Get("/quiz/{quizID}", pr.HandleGetQuiz)
		r.Post("/quiz/{quizID}/start", pr.HandleStartQuiz)
		r.Post("/quiz/{quizID}/submit", pr.HandleSubmitQuiz)
	})

	return r
}
