package accounts

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateAccount(ctx context.Context, account *Account) error
	GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, error)
	ListAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]Account, error)
	GetAccountByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*Account, error)
	UpdateAccountBalanceInTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, newBalance int64) error
}

type pgxRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &pgxRepository{db: db}
}

func (r *pgxRepository) CreateAccount(ctx context.Context, acc *Account) error {
	query := `
		INSERT INTO accounts (id, user_id, name, type, balance, currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(ctx, query,
		acc.ID,
		acc.UserID,
		acc.Name,
		acc.Type,
		acc.Balance,
		acc.Currency,
		acc.CreatedAt,
		acc.UpdatedAt,
	)
	return err
}

func (r *pgxRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	query := `
		SELECT id, user_id, name, type, balance, currency, created_at, updated_at
		FROM accounts
		WHERE id = $1`

	var acc Account
	err := r.db.QueryRow(ctx, query, id).Scan(
		&acc.ID,
		&acc.UserID,
		&acc.Name,
		&acc.Type,
		&acc.Balance,
		&acc.Currency,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &acc, nil
}

func (r *pgxRepository) ListAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]Account, error) {
	query := `
		SELECT id, user_id, name, type, balance, currency, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		err := rows.Scan(
			&acc.ID,
			&acc.UserID,
			&acc.Name,
			&acc.Type,
			&acc.Balance,
			&acc.Currency,
			&acc.CreatedAt,
			&acc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	return accounts, nil
}

func (r *pgxRepository) GetAccountByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*Account, error) {
	query := `
		SELECT id, user_id, name, type, balance, currency, created_at, updated_at
		FROM accounts
		WHERE id = $1
		FOR UPDATE`

	var acc Account
	err := tx.QueryRow(ctx, query, id).Scan(
		&acc.ID,
		&acc.UserID,
		&acc.Name,
		&acc.Type,
		&acc.Balance,
		&acc.Currency,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &acc, nil
}

func (r *pgxRepository) UpdateAccountBalanceInTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, newBalance int64) error {
	query := `
		UPDATE accounts
		SET balance = $1, updated_at = NOW()
		WHERE id = $2`

	tag, err := tx.Exec(ctx, query, newBalance, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errors.New("account not found or not updated")
	}

	return nil
}
