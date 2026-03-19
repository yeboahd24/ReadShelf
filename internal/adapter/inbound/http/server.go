package http

import (
	"net/http"

	"github.com/dominic/readshelf/internal/adapter/inbound/http/handler"
	"github.com/dominic/readshelf/internal/adapter/inbound/http/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	jwtSecret string,
	authH *handler.AuthHandler,
	bookH *handler.BookHandler,
	annotationH *handler.AnnotationHandler,
	recallH *handler.RecallHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api", func(api chi.Router) {
		// Public auth routes.
		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/register", authH.Register)
			auth.Post("/login", authH.Login)
			auth.Post("/refresh", authH.Refresh)
		})

		// Protected routes.
		api.Group(func(protected chi.Router) {
			protected.Use(middleware.JWTAuth(jwtSecret))

			// Books.
			protected.Get("/books", bookH.List)
			protected.Post("/books", bookH.Upload)
			protected.Get("/books/{id}", bookH.Get)
			protected.Get("/books/{id}/url", bookH.GetURL)
			protected.Delete("/books/{id}", bookH.Delete)

			// Annotations.
			protected.Get("/books/{id}/annotations", annotationH.ListByBook)
			protected.Post("/books/{id}/annotations", annotationH.Create)
			protected.Delete("/annotations/{id}", annotationH.Delete)
			protected.Patch("/annotations/{id}/note", annotationH.UpdateNote)
			protected.Get("/annotations/search", annotationH.Search)

			// AI Recall.
			protected.Post("/recall", recallH.Query)
		})
	})

	return r
}
