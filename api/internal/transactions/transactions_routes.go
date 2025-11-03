package transactions

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/transactions", h.HandleCreateTransaction)

	r.Get("/accounts/{accountID}/transactions", h.HandleListTransactionsByAccount)
}
