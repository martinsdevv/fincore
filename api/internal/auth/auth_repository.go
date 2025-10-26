package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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
	query := `INSERT INTO users (id, first_name, last_name, email, password)
              VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
	)
	return err
}

func (r *pgxRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, first_name, last_name, email, password
              FROM users
              WHERE email = $1`

	var user User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
	)

	if err != nil {
		// Se o erro for "nenhuma linha encontrada", retornamos (nil, nil).
		// O serviço vai interpretar isso como "usuário não existe".
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		// Se for qualquer outro erro de banco, nós o retornamos.
		return nil, err
	}

	return &user, nil
}
