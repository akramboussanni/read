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

	// Apply auth and admin-only middleware
	middleware.AddAuth(r, ar.Repos.User, ar.Repos.Token)
	r.Use(middleware.RequireEmailConfirmed)
	// assuming RequireAdmin is applied correctly according to codebase (looks like it's a func taking chi.Router based on `middleware.RequireAdmin(r)`)
	middleware.RequireAdmin(r)
	middleware.AddRatelimit(r, 30, 1*time.Minute)

	// User management
	r.Get("/users", ar.HandleListUsers)
	r.Get("/users/{userID}", ar.HandleGetUserDetail)
	r.Post("/users/{userID}/password", ar.HandleChangePassword)
	r.Delete("/users/{userID}", ar.HandleDeleteUser)

	// Course management (replaces old quiz admin)
	r.Route("/courses", func(r chi.Router) {
		r.Get("/", ar.HandleListCourses)
		r.Post("/", ar.HandleCreateCourse)
		r.Post("/auto-generate", ar.HandleAutoGenerateCourse)
		r.Get("/templates", ar.HandleListTemplates)
		r.Post("/from-template", ar.HandleCreateFromTemplate)
		r.Get("/{courseID}", ar.HandleGetCourse)
		r.Put("/{courseID}", ar.HandleUpdateCourse)
		r.Delete("/{courseID}", ar.HandleDeleteCourse)

		// Node management
		r.Post("/{courseID}/nodes", ar.HandleAddNode)
		r.Put("/nodes/{nodeID}", ar.HandleUpdateNode)
		r.Delete("/nodes/{nodeID}", ar.HandleDeleteNode)

		// Edge management
		r.Post("/{courseID}/edges", ar.HandleAddEdge)
		r.Delete("/edges/{edgeID}", ar.HandleDeleteEdge)
	})

	// Deck listing
	r.Get("/decks", ar.HandleListDecks)

	// System statistics
	r.Get("/stats/overview", ar.HandleGetSystemStats)

	return r
}
