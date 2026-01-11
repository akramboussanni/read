package quiz

import (
	"net/http"
	"sync"
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
	DeckService *quizpkg.DeckService
	Repos       *repo.Repos
	// Simple in-memory session store for demo purposes
	// In production, this would be a database or Redis
	sessions   map[string]*quizpkg.Quiz
	sessionMux sync.RWMutex
}

func NewQuizRouter(repos *repo.Repos, quizService *quizpkg.QuizService) http.Handler {
	qr := &QuizRouter{
		UserRepo:    repos.User,
		TokenRepo:   repos.Token,
		QuizService: quizService,
		DeckService: quizpkg.NewDeckService(repos),
		Repos:       repos,
		sessions:    make(map[string]*quizpkg.Quiz),
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

		r.Post("/", qr.HandleCreateQuiz) // Create quiz
		r.Post("/{quizID}/answers", qr.HandleSubmitAnswer)
		r.Get("/{quizID}/progress", qr.HandleGetQuizProgress)
	})

	// Question management (authenticated)
	r.Group(func(r chi.Router) {
		middleware.AddAuth(r, qr.UserRepo, qr.TokenRepo)
		middleware.AddRatelimit(r, 20, 1*time.Minute)

		r.Post("/questions", qr.HandleCreateQuestion)
	})

	return r
}
