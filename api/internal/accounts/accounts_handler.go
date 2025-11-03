package accounts

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/martinsdevv/fincore/internal/auth"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	service  Service
	validate *validator.Validate
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("Falha ao escrever resposta JSON")
	}
}

func (h *Handler) getUserIDFromContext(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		log.Error().Msg("UserID não encontrado no contexto, middleware mal configurado")
		return "", false
	}
	return userID, true
}

func (h *Handler) HandleCreateAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDFromContext(r)
	if !ok {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}

	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed: " + err.Error()})
		return
	}

	accountResp, err := h.service.CreateAccount(r.Context(), req, userID)
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create account"})
		return
	}

	h.writeJSON(w, http.StatusCreated, accountResp)
}

func (h *Handler) HandleGetAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDFromContext(r)
	if !ok {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}

	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing account ID"})
		return
	}

	accountResp, err := h.service.GetAccount(r.Context(), accountID, userID)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			h.writeJSON(w, http.StatusNotFound, map[string]string{"error": "account not found"})
			return
		}
		if errors.Is(err, ErrForbidden) {
			h.writeJSON(w, http.StatusForbidden, map[string]string{"error": "you do not have permission to view this account"})
			return
		}
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve account"})
		return
	}

	h.writeJSON(w, http.StatusOK, accountResp)
}

func (h *Handler) HandleListAccounts(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDFromContext(r)
	if !ok {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}

	accounts, err := h.service.ListAccounts(r.Context(), userID)
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve accounts"})
		return
	}

	h.writeJSON(w, http.StatusOK, accounts)
}
