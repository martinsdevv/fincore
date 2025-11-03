package transactions

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/fincore/internal/accounts"
	"github.com/stretchr/testify/assert"
)

// IDs globais para facilitar os testes
var (
	testUserID    = uuid.New()
	testAccountID = uuid.New()
	otherUserID   = uuid.New()
)

// Helper para criar um novo service com mocks
func setupServiceMocks() (*MockAccountsRepository, *MockTransactionsRepository, Service) {
	mockAccRepo := new(MockAccountsRepository)
	mockTxRepo := new(MockTransactionsRepository)
	// Removemos o prefixo "transactions." de NewService e Service
	service := NewService(nil, mockAccRepo, mockTxRepo)
	return mockAccRepo, mockTxRepo, service
}

func TestListTransactionsByAccount_Success(t *testing.T) {
	mockAccRepo, mockTxRepo, service := setupServiceMocks()
	ctx := context.Background()

	// 1. Setup: Mock da checagem de segurança (GetAccountByID)
	mockAccount := &accounts.Account{
		ID:     testAccountID,
		UserID: testUserID, // O usuário é o dono
	}
	mockAccRepo.On("GetAccountByID", ctx, testAccountID).Return(mockAccount, nil)

	// 2. Setup: Mock do retorno das transações (removemos prefixo de Transaction)
	mockTxs := []Transaction{
		{ID: uuid.New(), AccountID: testAccountID, Amount: 1000, Type: "income", TransactionDate: time.Now(), CreatedAt: time.Now()},
	}
	mockTxRepo.On("ListTransactionsByAccountID", ctx, testAccountID).Return(mockTxs, nil)

	// 3. Execução
	resp, err := service.ListTransactionsByAccount(ctx, testAccountID.String(), testUserID.String())

	// 4. Assertivas
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, int64(1000), resp[0].Amount)
	mockAccRepo.AssertExpectations(t)
	mockTxRepo.AssertExpectations(t)
}

func TestListTransactionsByAccount_Forbidden(t *testing.T) {
	mockAccRepo, _, service := setupServiceMocks()
	ctx := context.Background()

	// 1. Setup: Mock da checagem de segurança (GetAccountByID)
	mockAccount := &accounts.Account{
		ID:     testAccountID,
		UserID: otherUserID,
	}
	mockAccRepo.On("GetAccountByID", ctx, testAccountID).Return(mockAccount, nil)

	// 2. Execução
	resp, err := service.ListTransactionsByAccount(ctx, testAccountID.String(), testUserID.String())

	// 3. Assertivas
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrForbidden) // Removemos prefixo de ErrForbidden
	mockAccRepo.AssertExpectations(t)
}

func TestListTransactionsByAccount_AccountNotFound(t *testing.T) {
	mockAccRepo, _, service := setupServiceMocks()
	ctx := context.Background()

	// 1. Setup: Mock da checagem (GetAccountByID)
	mockAccRepo.On("GetAccountByID", ctx, testAccountID).Return(nil, nil)

	// 2. Execução
	resp, err := service.ListTransactionsByAccount(ctx, testAccountID.String(), testUserID.String())

	// 3. Assertivas
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrAccountNotFound) // Removemos prefixo de ErrAccountNotFound
	mockAccRepo.AssertExpectations(t)
}
