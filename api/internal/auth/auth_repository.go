package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type pgxRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &pgxRepository{db: db}
}

func (r *pgxRepository) CreateUser(ctx context.Context, user *User) error {
	// TODO: Adicionar lógica para inserir usuário
	return nil
}

func (r *pgxRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	// TODO: Adicionar lógica para buscar usuário
	return nil, nil
}
