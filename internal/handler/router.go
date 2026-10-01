package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/mbeltran93/taskflow-go/internal/auth"
	"github.com/mbeltran93/taskflow-go/internal/middleware"
)

// Handlers bundles everything the router needs to wire routes to behavior.
type Handlers struct {
	Auth     *AuthHandler
	Users    *UserHandler
	Projects *ProjectHandler
	Tasks    *TaskHandler
}

// NewRouter builds the full chi.Router for the API: global middleware,
// public routes, and JWT-protected write routes.
func NewRouter(h Handlers, jwtManager *auth.JWTManager) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(middleware.RequestLogger)
	r.Use(chimw.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/openapi.yaml", ServeOpenAPISpec)
	r.Get("/swagger", ServeSwaggerUI)
	r.Get("/swagger/", ServeSwaggerUI)

	requireAuth := middleware.RequireAuth(jwtManager)

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/login", h.Auth.Login)

		r.Route("/users", func(r chi.Router) {
			r.Get("/", h.Users.List)
			r.Post("/", h.Users.Register)
			r.Get("/{id}", h.Users.Get)
		})

		r.Route("/projects", func(r chi.Router) {
			r.Get("/", h.Projects.List)
			r.Get("/{id}", h.Projects.Get)

			r.Group(func(r chi.Router) {
				r.Use(requireAuth)
				r.Post("/", h.Projects.Create)
				r.Put("/{id}", h.Projects.Update)
				r.Delete("/{id}", h.Projects.Delete)
			})
		})

		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", h.Tasks.List)
			r.Get("/{id}", h.Tasks.Get)

			r.Group(func(r chi.Router) {
				r.Use(requireAuth)
				r.Post("/", h.Tasks.Create)
				r.Put("/{id}", h.Tasks.Update)
				r.Patch("/{id}/status", h.Tasks.UpdateStatus)
				r.Delete("/{id}", h.Tasks.Delete)
			})
		})
	})

	return r
}
