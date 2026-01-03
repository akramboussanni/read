package routes

import (
	"net/http"

	"github.com/akramboussanni/gocode/internal/api"
	"github.com/akramboussanni/gocode/internal/api/routes/admin"
	"github.com/akramboussanni/gocode/internal/api/routes/auth"
	"github.com/akramboussanni/gocode/internal/api/routes/progression"
	"github.com/akramboussanni/gocode/internal/api/routes/quiz"
	"github.com/akramboussanni/gocode/internal/middleware"
	quizpkg "github.com/akramboussanni/gocode/internal/quiz"
	"github.com/akramboussanni/gocode/internal/repo"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func SetupRouter(repos *repo.Repos) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORSHeaders)

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("github.com/akramboussanni/gocode"))
	})

	api.AddSwaggerRoutes(r)

	// Initialize quiz service
	quizService := quizpkg.NewQuizService(repos)

	// Mount routers
	r.Mount("/auth", auth.NewAuthRouter(repos.User, repos.Token, repos.Lockout))
	r.Mount("/progression", progression.NewProgressionRouter(repos, quizService))
	r.Mount("/quizzes", quiz.NewQuizRouter(repos, quizService))

	// Admin routes (requires authentication and admin role)
	r.Route("/admin", func(r chi.Router) {
		middleware.AddAuth(r, repos.User, repos.Token)
		r.Mount("/", admin.NewAdminRouter(repos).Routes())
	})

	return r
}
