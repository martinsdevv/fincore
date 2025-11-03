package transactions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateTransactionInTx(ctx context.Context, tx pgx.Tx, tr *Transaction) error

	ListTransactionsByAccountID(ctx context.Context, accountID uuid.UUID) ([]Transaction, error)
}

type pgxRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &pgxRepository{db: db}
}

func (r *pgxRepository) CreateTransactionInTx(ctx context.Context, tx pgx.Tx, tr *Transaction) error {
	query := `
		INSERT INTO transactions (id, account_id, type, amount, description, category, transaction_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := tx.Exec(ctx, query,
		tr.ID,
		tr.AccountID,
		tr.Type,
		tr.Amount,
		tr.Description,
		tr.Category,
		tr.TransactionDate,
		tr.CreatedAt,
	)
	return err
}

func (r *pgxRepository) ListTransactionsByAccountID(ctx context.Context, accountID uuid.UUID) ([]Transaction, error) {
	query := `
		SELECT id, account_id, type, amount, description, category, transaction_date, created_at
		FROM transactions
		WHERE account_id = $1
		ORDER BY transaction_date DESC, created_at DESC`

	rows, err := r.db.Query(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tr Transaction
		err := rows.Scan(
			&tr.ID,
			&tr.AccountID,
			&tr.Type,
			&tr.Amount,
			&tr.Description,
			&tr.Category,
			&tr.TransactionDate,
			&tr.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tr)
	}

	return transactions, nil
}
