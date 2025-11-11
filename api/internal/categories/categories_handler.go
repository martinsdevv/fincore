package categories

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/martinsdevv/fincore/internal/auth" // Para o UserContextKey
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

// --- CRUD Handlers ---

func (h *Handler) HandleCreateCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDFromContext(r)
	if !ok {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}

	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed: " + err.Error()})
		return
	}

	resp, err := h.service.CreateCategory(r.Context(), req, userID)
	if err != nil {
		// TODO: Mapear o erro de 'uq_user_category_name' para 409 Conflict
		log.Error().Err(err).Msg("Erro não mapeado ao criar categoria")
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create category"})
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) HandleGetCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDFromContext(r)
	if !ok {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}

	categoryID := chi.URLParam(r, "categoryID")
	if categoryID == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing category ID"})
		return
	}

	resp, err := h.service.GetCategory(r.Context(), categoryID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrCategoryNotFound):
			h.writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrForbidden):
			h.writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrInvalidUUID):
			h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			log.Error().Err(err).Str("categoryID", categoryID).Msg("Erro não mapeado ao buscar categoria")
			h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve category"})
		}
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleListCategories(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDFromContext(r)
	if !ok {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}

	categories, err := h.service.ListCategories(r.Context(), userID)
	if err != nil {
		// Listar geralmente não dá erros 404/403, só 500 ou 400 se tiver validação de ID
		log.Error().Err(err).Str("userID", userID).Msg("Erro não mapeado ao listar categorias")
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve categories"})
		return
	}

	if len(categories) == 0 {
		h.writeJSON(w, http.StatusOK, []CategoryResponse{}) // Retorna lista vazia em vez de nulo
		return
	}

	h.writeJSON(w, http.StatusOK, categories)
}

func (h *Handler) HandleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDFromContext(r)
	if !ok {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}

	categoryID := chi.URLParam(r, "categoryID")
	if categoryID == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing category ID"})
		return
	}

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed: " + err.Error()})
		return
	}

	resp, err := h.service.UpdateCategory(r.Context(), categoryID, req, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrCategoryNotFound):
			h.writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrForbidden):
			h.writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrInvalidUUID):
			h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		// TODO: Mapear o erro de 'uq_user_category_name' para 409 Conflict
		default:
			log.Error().Err(err).Str("categoryID", categoryID).Msg("Erro não mapeado ao atualizar categoria")
			h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update category"})
		}
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDFromContext(r)
	if !ok {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}

	categoryID := chi.URLParam(r, "categoryID")
	if categoryID == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing category ID"})
		return
	}

	err := h.service.DeleteCategory(r.Context(), categoryID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrCategoryNotFound):
			h.writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrForbidden):
			h.writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrInvalidUUID):
			h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			log.Error().Err(err).Str("categoryID", categoryID).Msg("Erro não mapeado ao deletar categoria")
			h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete category"})
		}
		return
	}

	h.writeJSON(w, http.StatusNoContent, nil) // 204 No Content é a resposta padrão para DELETE
}

// --- RegisterRoutes ---

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/categories", h.HandleCreateCategory)
	r.Get("/categories", h.HandleListCategories)
	r.Get("/categories/{categoryID}", h.HandleGetCategory)
	r.Put("/categories/{categoryID}", h.HandleUpdateCategory)
	r.Delete("/categories/{categoryID}", h.HandleDeleteCategory)
}
