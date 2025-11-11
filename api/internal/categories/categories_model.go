package categories

import (
	"time"

	"github.com/google/uuid"
)

// Category é a struct do banco de dados
type Category struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"` // Dono da categoria
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCategoryRequest é o que esperamos do body do JSON para criar
type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

// UpdateCategoryRequest é o que esperamos do body do JSON para atualizar
type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

// CategoryResponse é o que retornamos para o frontend
type CategoryResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Helper para converter o modelo do DB para a resposta da API
func toCategoryResponse(cat *Category) *CategoryResponse {
	return &CategoryResponse{
		ID:        cat.ID,
		UserID:    cat.UserID,
		Name:      cat.Name,
		CreatedAt: cat.CreatedAt,
	}
}
