package routes

import (
	"net/http"

	"github.com/akramboussanni/gocode/internal/api"
	"github.com/akramboussanni/gocode/internal/api/routes/admin"
	"github.com/akramboussanni/gocode/internal/api/routes/auth"
	"github.com/akramboussanni/gocode/internal/api/routes/classroom"
	"github.com/akramboussanni/gocode/internal/api/routes/course"
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

	// Initialize services
	quizService := quizpkg.NewQuizService(repos)
	deckService := quizpkg.NewDeckService(repos)

	// Mount routers
	r.Mount("/auth", auth.NewAuthRouter(repos.User, repos.Token, repos.Lockout))
	r.Mount("/courses", course.NewCourseRouter(repos, quizService))
	r.Mount("/onboarding", course.NewOnboardingRouter(repos, quizService))
	r.Mount("/classroom", classroom.NewClassroomRouter(repos))
	r.Mount("/quiz", quiz.NewQuizRouter(repos, quizService))

	// Admin routes (requires authentication and admin role)
	r.Mount("/admin", admin.NewAdminRouter(repos, quizService, deckService).Routes())

	return r
}
