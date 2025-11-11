package transactions

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TypeIncome  TransactionType = "income"
	TypeExpense TransactionType = "expense"
)

// Transaction (Struct do Banco)
type Transaction struct {
	ID              uuid.UUID       `json:"id"`
	AccountID       uuid.UUID       `json:"account_id"`
	Type            TransactionType `json:"type"`
	Amount          int64           `json:"amount"`
	Description     string          `json:"description"`
	CategoryID      *uuid.UUID      `json:"category_id,omitempty"`
	TransactionDate time.Time       `json:"transaction_date"`
	CreatedAt       time.Time       `json:"created_at"`
}

// CreateTransactionRequest (DTO de Criação)
type CreateTransactionRequest struct {
	AccountID       string `json:"account_id" validate:"required,uuid"`
	Type            string `json:"type" validate:"required,oneof=income expense"`
	Amount          int64  `json:"amount" validate:"required,gt=0"`
	Description     string `json:"description" validate:"required,max=255"`
	CategoryID      string `json:"category_id" validate:"required,uuid"`
	TransactionDate string `json:"transaction_date" validate:"omitempty,datetime=2006-01-02"`
}

// TransactionResponse (DTO de Resposta)
type TransactionResponse struct {
	ID              uuid.UUID       `json:"id"`
	AccountID       uuid.UUID       `json:"account_id"`
	Type            TransactionType `json:"type"`
	Amount          int64           `json:"amount"`
	Description     string          `json:"description"`
	CategoryID      *uuid.UUID      `json:"category_id,omitempty"`
	CategoryName    *string         `json:"category_name,omitempty"` // <-- CAMPO ADICIONADO
	TransactionDate time.Time       `json:"transaction_date"`
	CreatedAt       time.Time       `json:"created_at"`
}

type ListTransactionsRow struct {
	ID              uuid.UUID       `json:"id"`
	AccountID       uuid.UUID       `json:"account_id"`
	Type            TransactionType `json:"type"`
	Amount          int64           `json:"amount"`
	Description     string          `json:"description"`
	CategoryID      *uuid.UUID      `json:"category_id,omitempty"`
	TransactionDate time.Time       `json:"transaction_date"`
	CreatedAt       time.Time       `json:"created_at"`
	CategoryName    sql.NullString  `json:"category_name"`
}
