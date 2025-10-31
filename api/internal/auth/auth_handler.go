package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

type contextKey string

const UserContextKey = contextKey("userID")

type Handler struct {
	service   Service
	validate  *validator.Validate
	jwtSecret string
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("Falha ao escrever resposta JSON")
	}
}

func NewHandler(service Service, jwtSecret string) *Handler {
	return &Handler{
		service:   service,
		validate:  validator.New(validator.WithRequiredStructEnabled()),
		jwtSecret: jwtSecret,
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

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization header format"})
			return
		}
		tokenString := headerParts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(h.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			log.Warn().Err(err).Msg("Invalid token attempt")
			h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid user ID in token"})
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserContextKey).(string)
	if !ok {
		log.Error().Msg("UserID não encontrado no contexto, middleware mal configurado")
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	userResponse, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			h.writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}

		log.Error().Err(err).Str("userID", userID).Msg("Falha ao buscar perfil do usuário")
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve user profile"})
		return
	}

	h.writeJSON(w, http.StatusOK, userResponse)
}
