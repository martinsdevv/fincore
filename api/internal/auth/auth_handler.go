package auth

import (
	"net/http"

	"github.com/go-playground/validator/v10"
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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	// TODO: 1. Decodificar JSON
	// TODO: 2. Validar struct (h.validate)
	// TODO: 3. Chamar h.service.Register
	// TODO: 4. Escrever resposta JSON (sucesso ou erro)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "register endpoint"}`))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	// TODO: 1. Decodificar JSON
	// TODO: 2. Validar struct
	// TODO: 3. Chamar h.service.Login
	// TODO: 4. Escrever resposta JSON
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "login endpoint"}`))
}
