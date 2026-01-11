package quiz

import (
	"net/http"
	"time"

	"github.com/akramboussanni/gocode/internal/middleware"
	quizpkg "github.com/akramboussanni/gocode/internal/quiz"
	"github.com/akramboussanni/gocode/internal/repo"
	"github.com/go-chi/chi/v5"
)

type QuizRouter struct {
	UserRepo    *repo.UserRepo
	TokenRepo   *repo.TokenRepo
	QuizService *quizpkg.QuizService
	Repos       *repo.Repos
}

func NewQuizRouter(repos *repo.Repos, quizService *quizpkg.QuizService) http.Handler {
	qr := &QuizRouter{
		UserRepo:    repos.User,
		TokenRepo:   repos.Token,
		QuizService: quizService,
		Repos:       repos,
	}
	r := chi.NewRouter()

	r.Use(middleware.MaxBytesMiddleware(1 << 20))

	// Authenticated deck/category browsing (must come before /{quizID} route)
	r.Group(func(r chi.Router) {
		middleware.AddAuth(r, qr.UserRepo, qr.TokenRepo)
		middleware.AddRatelimit(r, 30, 1*time.Minute)

		r.Get("/decks", qr.HandleListDecks)
		r.Get("/decks/{deckID}/categories", qr.HandleGetCategories)
	})

	// Public quiz browsing (light rate limit)
	r.Group(func(r chi.Router) {
		middleware.AddRatelimit(r, 60, 1*time.Minute)
		r.Get("/", qr.HandleListQuizzes)
		r.Get("/{quizID}", qr.HandleGetQuiz)
	})

	// Authenticated quiz operations
	r.Group(func(r chi.Router) {
		middleware.AddAuth(r, qr.UserRepo, qr.TokenRepo)
		middleware.AddRatelimit(r, 30, 1*time.Minute)

		r.Get("/my", qr.HandleGetMyQuizzes)
		r.Post("/", qr.HandleCreateQuiz)
		r.Post("/generate", qr.HandleGenerateQuiz)
		r.Put("/{quizID}", qr.HandleUpdateQuiz)
		r.Delete("/{quizID}", qr.HandleDeleteQuiz)
		r.Post("/{quizID}/start", qr.HandleStartQuiz)
		r.Post("/{quizID}/submit", qr.HandleSubmitQuiz)
	})

	// Question management (authenticated)
	r.Group(func(r chi.Router) {
		middleware.AddAuth(r, qr.UserRepo, qr.TokenRepo)
		middleware.AddRatelimit(r, 20, 1*time.Minute)

		r.Post("/questions", qr.HandleCreateQuestion)
	})

	return r
}
