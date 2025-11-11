package categories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository (interface)
// (Isso deve estar no seu model.go ou em um arquivo de interface,
// mas vou colocar aqui para ficar completo)
type Repository interface {
	CreateCategory(ctx context.Context, cat *Category) error
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error)
	ListCategoriesByUserID(ctx context.Context, userID uuid.UUID) ([]Category, error)
	UpdateCategory(ctx context.Context, cat *Category) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}

// pgxRepository (implementação)
type pgxRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &pgxRepository{db: db}
}

func (r *pgxRepository) CreateCategory(ctx context.Context, cat *Category) error {
	query := `
		INSERT INTO categories (id, user_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, query,
		cat.ID,
		cat.UserID,
		cat.Name,
		cat.CreatedAt,
		cat.UpdatedAt,
	)
	return err
}

func (r *pgxRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error) {
	query := `
		SELECT id, user_id, name, created_at, updated_at
		FROM categories
		WHERE id = $1`

	var cat Category
	err := r.db.QueryRow(ctx, query, id).Scan(
		&cat.ID,
		&cat.UserID,
		&cat.Name,
		&cat.CreatedAt,
		&cat.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // O serviço vai tratar nil como "Not Found"
		}
		return nil, err
	}
	return &cat, nil
}

func (r *pgxRepository) ListCategoriesByUserID(ctx context.Context, userID uuid.UUID) ([]Category, error) {
	query := `
		SELECT id, user_id, name, created_at, updated_at
		FROM categories
		WHERE user_id = $1
		ORDER BY name ASC` // Ordenar por nome faz sentido

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		err := rows.Scan(
			&cat.ID,
			&cat.UserID,
			&cat.Name,
			&cat.CreatedAt,
			&cat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

func (r *pgxRepository) UpdateCategory(ctx context.Context, cat *Category) error {
	query := `
		UPDATE categories
		SET name = $1, updated_at = $2
		WHERE id = $3`

	tag, err := r.db.Exec(ctx, query, cat.Name, cat.UpdatedAt, cat.ID)
	if err != nil {
		return err
	}
	// Se o serviço checou a permissão, o único erro aqui
	// seria se o ID não existisse, o que é um caso raro.
	if tag.RowsAffected() == 0 {
		return errors.New("category not found or not updated")
	}
	return nil
}

func (r *pgxRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM categories WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("category not found") // O serviço vai tratar isso
	}
	return nil
}
