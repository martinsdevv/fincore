package auth

import (
	"net/http"

	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	service  Service
	validate *validator.Validate
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("Falha ao escrever resposta JSON")
	}
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed: " + err.Error()})
		return
	}

	if err := h.service.Register(r.Context(), req); err != nil {
		// TODO: Mapear erros do serviço (ex: email duplicado -> 409 Conflict)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to register user"})
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]string{"message": "user registered successfully"})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed: " + err.Error()})
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		// TODO: Mapear erros do serviço (ex: credenciais erradas -> 401 Unauthorized)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "login failed"}) // Provisório
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}
