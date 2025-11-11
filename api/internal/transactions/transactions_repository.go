package transactions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateTransactionInTx(ctx context.Context, tx pgx.Tx, tr *Transaction) error
	// --- MUDANÇA NA ASSINATURA ---
	// Agora retorna a struct do JOIN
	ListTransactionsByAccountID(ctx context.Context, accountID uuid.UUID) ([]ListTransactionsRow, error)
}

type pgxRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &pgxRepository{db: db}
}

func (r *pgxRepository) CreateTransactionInTx(ctx context.Context, tx pgx.Tx, tr *Transaction) error {
	// (Esta função continua a mesma)
	query := `
		INSERT INTO transactions (id, account_id, type, amount, description, category_id, transaction_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := tx.Exec(ctx, query,
		tr.ID,
		tr.AccountID,
		tr.Type,
		tr.Amount,
		tr.Description,
		tr.CategoryID,
		tr.TransactionDate,
		tr.CreatedAt,
	)
	return err
}

// --- GRANDES MUDANÇAS AQUI ---
func (r *pgxRepository) ListTransactionsByAccountID(ctx context.Context, accountID uuid.UUID) ([]ListTransactionsRow, error) {
	// --- MUDANÇA NA QUERY (LEFT JOIN) ---
	query := `
		SELECT
			t.id,
			t.account_id,
			t.type,
			t.amount,
			t.description,
			t.category_id,
			t.transaction_date,
			t.created_at,
			c.name AS category_name
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE t.account_id = $1
		ORDER BY t.transaction_date DESC, t.created_at DESC`

	rows, err := r.db.Query(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// --- MUDANÇA NO SCAN ---
	var transactions []ListTransactionsRow // <-- Usa a nova struct
	for rows.Next() {
		var row ListTransactionsRow // <-- Usa a nova struct
		err := rows.Scan(
			&row.ID,
			&row.AccountID,
			&row.Type,
			&row.Amount,
			&row.Description,
			&row.CategoryID,
			&row.TransactionDate,
			&row.CreatedAt,
			&row.CategoryName, // <-- Escaneia o novo campo
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, row)
	}

	return transactions, nil
}
