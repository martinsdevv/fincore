package transactions

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/martinsdevv/fincore/internal/auth" // Precisamos da chave do contexto
	"github.com/rs/zerolog/log"
)

type Handler struct {
	service  Service
	validate *validator.Validate
}

func NewHandler(service Service) *Handler {
	v := validator.New(validator.WithRequiredStructEnabled())
	// TODO: Registrar validação customizada para 'oneof=income expense' se necessário

	return &Handler{
		service:  service,
		validate: v,
	}
}

// writeJSON (helper)
func (h *Handler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("Falha ao escrever resposta JSON")
	}
}

// getUserIDFromContext (helper)
func (h *Handler) getUserIDFromContext(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		log.Error().Msg("UserID não encontrado no contexto, middleware mal configurado")
		return "", false
	}
	return userID, true
}

func (h *Handler) HandleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDFromContext(r)
	if !ok {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed: " + err.Error()})
		return
	}

	resp, err := h.service.CreateTransaction(r.Context(), req, userID)
	if err != nil {
		// Mapeamento de Erros de Negócio para Status HTTP
		switch {
		case errors.Is(err, ErrInsufficientFunds):
			// 422 Unprocessable Entity é bom para falhas de lógica de negócio
			h.writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrForbidden):
			h.writeJSON(w, http.StatusForbidden, map[string]string{"error": "you do not have permission for this account"})
		case errors.Is(err, ErrAccountNotFound):
			h.writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			log.Error().Err(err).Msg("Erro não mapeado ao criar transação")
			h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create transaction"})
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) HandleListTransactionsByAccount(w http.ResponseWriter, r *http.Request) {
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

	transactions, err := h.service.ListTransactionsByAccount(r.Context(), accountID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			h.writeJSON(w, http.StatusForbidden, map[string]string{"error": "you do not have permission to view this account"})
		case errors.Is(err, ErrAccountNotFound):
			h.writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			log.Error().Err(err).Str("accountID", accountID).Msg("Erro ao listar transações")
			h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve transactions"})
		}
		return
	}

	if len(transactions) == 0 {
		h.writeJSON(w, http.StatusOK, []TransactionResponse{})
		return
	}

	h.writeJSON(w, http.StatusOK, transactions)
}
