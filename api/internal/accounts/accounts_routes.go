package accounts

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/accounts", h.HandleCreateAccount)
	r.Get("/accounts", h.HandleListAccounts)
	r.Get("/accounts/{accountID}", h.HandleGetAccount)
	// r.Put("/accounts/{accountID}", h.HandleUpdateAccount)
	// r.Delete("/accounts/{accountID}", h.HandleDeleteAccount)
}
