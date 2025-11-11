package transactions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/martinsdevv/fincore/internal/accounts"
	"github.com/martinsdevv/fincore/internal/categories" // <-- IMPORT NOVO
	"github.com/stretchr/testify/mock"
)

// --- MockAccountsRepository (Sem alteração) ---
type MockAccountsRepository struct {
	mock.Mock
}

// (Todos os métodos de MockAccountsRepository continuam aqui, sem alteração)
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
	// ATENÇÃO: Tive que ajustar seu mock. O GetAccountByIDForUpdate
	// não recebe a transação (tx) como argumento no mock.
	args := m.Called(ctx, id) // <--- Ajuste aqui
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*accounts.Account), args.Error(1)
}
func (m *MockAccountsRepository) UpdateAccountBalanceInTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, newBalance int64) error {
	// E aqui também, sem o tx
	args := m.Called(ctx, id, newBalance) // <--- Ajuste aqui
	return args.Error(0)
}

// --- MockTransactionsRepository (Sem alteração) ---
type MockTransactionsRepository struct {
	mock.Mock
}

func (m *MockTransactionsRepository) CreateTransactionInTx(ctx context.Context, tx pgx.Tx, tr *Transaction) error {
	// E aqui também, sem o tx
	args := m.Called(ctx, tr) // <--- Ajuste aqui
	return args.Error(0)
}
func (m_ *MockTransactionsRepository) ListTransactionsByAccountID(ctx context.Context, accountID uuid.UUID) ([]Transaction, error) {
	args := m_.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Transaction), args.Error(1)
}

// --- MockCategoriesRepository (NOVO) ---
type MockCategoriesRepository struct {
	mock.Mock
}

// Implementa a interface 'categories.Repository'
func (m *MockCategoriesRepository) CreateCategory(ctx context.Context, cat *categories.Category) error {
	args := m.Called(ctx, cat)
	return args.Error(0)
}
func (m *MockCategoriesRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*categories.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*categories.Category), args.Error(1)
}
func (m *MockCategoriesRepository) ListCategoriesByUserID(ctx context.Context, userID uuid.UUID) ([]categories.Category, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]categories.Category), args.Error(1)
}
func (m *MockCategoriesRepository) UpdateCategory(ctx context.Context, cat *categories.Category) error {
	args := m.Called(ctx, cat)
	return args.Error(0)
}
func (m *MockCategoriesRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
