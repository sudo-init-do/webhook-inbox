package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(h *Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", h.Health)

	r.Route("/api", func(api chi.Router) {
		api.Post("/endpoints", h.CreateEndpoint)
		api.Get("/messages", h.ListMessages)
		api.Get("/messages/{id}", h.GetMessage)
		api.Post("/messages/{id}/replay", h.ReplayMessage) // NEW
	})

	// Webhook receiver
	r.Post("/hooks/{token}", h.ReceiveHook)

	// Default
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	return r
}

// small chi param helper
func pathParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
