package quiz

import (
	"net/http"

	"github.com/akramboussanni/gocode/internal/middleware"
	quizpkg "github.com/akramboussanni/gocode/internal/quiz"
	"github.com/akramboussanni/gocode/internal/repo"
	"github.com/go-chi/chi/v5"
)

type QuizRouter struct {
	Repos       *repo.Repos
	QuizService *quizpkg.QuizService
}

func NewQuizRouter(repos *repo.Repos, quizService *quizpkg.QuizService) http.Handler {
	qr := &QuizRouter{
		Repos:       repos,
		QuizService: quizService,
	}

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		middleware.AddAuth(r, repos.User, repos.Token)
		r.Use(middleware.RequireEmailConfirmed)

		// Quiz CRUD
		r.Get("/", qr.ListQuizzes)
		r.Post("/", qr.CreateQuiz)
		r.Get("/my", qr.MyQuizzes)
		r.Get("/{quizID}", qr.GetQuiz)
		r.Put("/{quizID}", qr.UpdateQuiz)
		r.Delete("/{quizID}", qr.DeleteQuiz)

		// Deck browsing (for quiz creation UI)
		r.Get("/decks", qr.ListDecks)
		r.Get("/decks/{deckID}/categories", qr.GetDeckCategories)

		// Question preview (generate sample questions for manual review)
		r.Post("/preview", qr.PreviewQuestions)

		// Quiz execution
		r.Post("/start", qr.StartQuiz)
		r.Post("/answer", qr.SubmitAnswer)
		r.Post("/complete", qr.CompleteQuiz)
		r.Get("/attempt/{attemptID}", qr.GetAttempt)
		r.Get("/my/history", qr.MyHistory)
		r.Post("/nodes/complete", qr.CompleteNode)
	})

	return r
}
