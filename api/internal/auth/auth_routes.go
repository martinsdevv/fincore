package auth

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
}
