package transactions

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TypeIncome  TransactionType = "income"
	TypeExpense TransactionType = "expense"
)

type Transaction struct {
	ID              uuid.UUID       `json:"id"`
	AccountID       uuid.UUID       `json:"account_id"` // Chave estrangeira para a conta
	Type            TransactionType `json:"type"`       // "income" ou "expense"
	Amount          int64           `json:"amount"`     // Em centavos, SEMPRE positivo
	Description     string          `json:"description"`
	Category        string          `json:"category"` // Ex: "Salário", "Alimentação", "Lazer"
	TransactionDate time.Time       `json:"transaction_date"`
	CreatedAt       time.Time       `json:"created_at"`
}

type CreateTransactionRequest struct {
	AccountID       string `json:"account_id" validate:"required,uuid"`
	Type            string `json:"type" validate:"required,oneof=income expense"`
	Amount          int64  `json:"amount" validate:"required,gt=0"` // Deve ser maior que zero
	Description     string `json:"description" validate:"required,max=255"`
	Category        string `json:"category" validate:"required,max=100"`
	TransactionDate string `json:"transaction_date" validate:"omitempty,datetime=2006-01-02"` // Formato YYYY-MM-DD
}

type TransactionResponse struct {
	ID              uuid.UUID       `json:"id"`
	AccountID       uuid.UUID       `json:"account_id"`
	Type            TransactionType `json:"type"`
	Amount          int64           `json:"amount"`
	Description     string          `json:"description"`
	Category        string          `json:"category"`
	TransactionDate time.Time       `json:"transaction_date"`
	CreatedAt       time.Time       `json:"created_at"`
}
