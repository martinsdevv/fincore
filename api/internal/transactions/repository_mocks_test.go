package transactions // <-- Alterado

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/martinsdevv/fincore/internal/accounts"
	"github.com/stretchr/testify/mock"
)

// MockAccountsRepository (continua igual, pois mocka um pacote externo)
type MockAccountsRepository struct {
	mock.Mock
}

func (m *MockAccountsRepository) CreateAccount(ctx context.Context, acc *accounts.Account) error {
	args := m.Called(ctx, acc)
	return args.Error(0)
}

func (m *MockAccountsRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*accounts.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*accounts.Account), args.Error(1)
}

func (m *MockAccountsRepository) ListAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]accounts.Account, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]accounts.Account), args.Error(1)
}

func (m *MockAccountsRepository) GetAccountByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*accounts.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*accounts.Account), args.Error(1)
}

func (m *MockAccountsRepository) UpdateAccountBalanceInTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, newBalance int64) error {
	args := m.Called(ctx, id, newBalance)
	return args.Error(0)
}

// MockTransactionsRepository (mocka o repositório local)
type MockTransactionsRepository struct {
	mock.Mock
}

// Removemos o prefixo "transactions." de Repository e Transaction
func (m *MockTransactionsRepository) CreateTransactionInTx(ctx context.Context, tx pgx.Tx, tr *Transaction) error {
	args := m.Called(ctx, tr)
	return args.Error(0)
}

func (m_ *MockTransactionsRepository) ListTransactionsByAccountID(ctx context.Context, accountID uuid.UUID) ([]Transaction, error) {
	args := m_.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Transaction), args.Error(1)
}
