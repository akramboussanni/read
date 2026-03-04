package course

import (
	"net/http"

	"github.com/akramboussanni/gocode/internal/middleware"
	quizpkg "github.com/akramboussanni/gocode/internal/quiz"
	"github.com/akramboussanni/gocode/internal/repo"
	"github.com/go-chi/chi/v5"
)

type CourseRouter struct {
	repos       *repo.Repos
	quizService *quizpkg.QuizService
}

func NewCourseRouter(repos *repo.Repos, quizService *quizpkg.QuizService) http.Handler {
	cr := &CourseRouter{
		repos:       repos,
		quizService: quizService,
	}

	r := chi.NewRouter()

	// Public routes
	r.Get("/", cr.ListCourses)
	r.Get("/{courseID}", cr.GetCourse)

	// Authenticated routes
	r.Group(func(r chi.Router) {
		middleware.AddAuth(r, repos.User, repos.Token)
		r.Use(middleware.RequireEmailConfirmed)

		r.Post("/{courseID}/enroll", cr.Enroll)
		r.Get("/my/enrollments", cr.MyEnrollments)
		r.Put("/my/active-course", cr.SetActiveCourse)
		r.Get("/{courseID}/status", cr.GetCourseStatus)
	})

	return r
}

// ============================================================
// ONBOARDING ROUTE
// ============================================================

type OnboardingRouter struct {
	repos       *repo.Repos
	quizService *quizpkg.QuizService
}

func NewOnboardingRouter(repos *repo.Repos, quizService *quizpkg.QuizService) http.Handler {
	or := &OnboardingRouter{
		repos:       repos,
		quizService: quizService,
	}

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		middleware.AddAuth(r, or.repos.User, or.repos.Token)
		r.Use(middleware.RequireEmailConfirmed)

		r.Get("/courses", or.AvailableCourses)
		r.Post("/complete", or.CompleteOnboarding)
	})
	return r
}
