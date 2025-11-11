package transactions

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/martinsdevv/fincore/internal/accounts"
	"github.com/martinsdevv/fincore/internal/categories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// IDs globais para facilitar os testes
var (
	testUserID     = uuid.New()
	testAccountID  = uuid.New()
	testCategoryID = uuid.New() // <-- NOVO ID GLOBAL
	otherUserID    = uuid.New()
)

// Helper para criar um novo service com mocks
func setupServiceMocks() (*MockAccountsRepository, *MockTransactionsRepository, *MockCategoriesRepository, Service) {
	mockAccRepo := new(MockAccountsRepository)
	mockTxRepo := new(MockTransactionsRepository)
	mockCatRepo := new(MockCategoriesRepository) // <-- NOVO MOCK

	// <-- ASSINATURA ATUALIZADA
	service := NewService(nil, mockAccRepo, mockTxRepo, mockCatRepo)

	return mockAccRepo, mockTxRepo, mockCatRepo, service
}

// --- Testes de ListTransactionsByAccount (Atualizados) ---

func TestListTransactionsByAccount_Success(t *testing.T) {
	mockAccRepo, mockTxRepo, _, service := setupServiceMocks() // <-- MUDOU AQUI
	ctx := context.Background()

	// 1. Setup: Mock da checagem de segurança (GetAccountByID)
	mockAccount := &accounts.Account{
		ID:     testAccountID,
		UserID: testUserID, // O usuário é o dono
	}
	mockAccRepo.On("GetAccountByID", ctx, testAccountID).Return(mockAccount, nil)

	// 2. Setup: Mock do retorno das transações
	// <-- STRUCT ATUALIZADA (usando CategoryID)
	mockTxs := []Transaction{
		{
			ID:         uuid.New(),
			AccountID:  testAccountID,
			Amount:     1000,
			Type:       TypeIncome,
			CategoryID: &testCategoryID, // <-- MUDOU AQUI
		},
	}
	mockTxRepo.On("ListTransactionsByAccountID", ctx, testAccountID).Return(mockTxs, nil)

	// 3. Execução
	resp, err := service.ListTransactionsByAccount(ctx, testAccountID.String(), testUserID.String())

	// 4. Assertivas
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, int64(1000), resp[0].Amount)
	assert.Equal(t, &testCategoryID, resp[0].CategoryID) // <-- NOVO ASSERT
	mockAccRepo.AssertExpectations(t)
	mockTxRepo.AssertExpectations(t)
}

func TestListTransactionsByAccount_Forbidden(t *testing.T) {
	mockAccRepo, _, _, service := setupServiceMocks() // <-- MUDOU AQUI
	ctx := context.Background()

	// (Resto do teste igual ao original)
	mockAccount := &accounts.Account{
		ID:     testAccountID,
		UserID: otherUserID,
	}
	mockAccRepo.On("GetAccountByID", ctx, testAccountID).Return(mockAccount, nil)
	resp, err := service.ListTransactionsByAccount(ctx, testAccountID.String(), testUserID.String())
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrForbidden)
	mockAccRepo.AssertExpectations(t)
}

func TestListTransactionsByAccount_AccountNotFound(t *testing.T) {
	mockAccRepo, _, _, service := setupServiceMocks() // <-- MUDOU AQUI
	ctx := context.Background()

	// (Resto do teste igual ao original)
	mockAccRepo.On("GetAccountByID", ctx, testAccountID).Return(nil, nil)
	resp, err := service.ListTransactionsByAccount(ctx, testAccountID.String(), testUserID.String())
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrAccountNotFound)
	mockAccRepo.AssertExpectations(t)
}

// --- Testes de CreateTransaction (NOVOS) ---

func TestCreateTransaction_Success_Expense(t *testing.T) {
	mockAccRepo, mockTxRepo, mockCatRepo, service := setupServiceMocks()
	ctx := context.Background()

	req := CreateTransactionRequest{
		AccountID:  testAccountID.String(),
		Type:       string(TypeExpense),
		Amount:     1000,
		CategoryID: testCategoryID.String(),
	}

	// 1. Mock: GetAccountByIDForUpdate (para validar conta e saldo)
	mockAccount := &accounts.Account{
		ID:      testAccountID,
		UserID:  testUserID,
		Balance: 5000, // Saldo suficiente
	}
	// Note que o tx não é passado para o mock, como ajustamos em repository_mocks_test.go
	mockAccRepo.On("GetAccountByIDForUpdate", ctx, testAccountID).Return(mockAccount, nil)

	// 2. Mock: GetCategoryByID (para validar categoria)
	mockCategory := &categories.Category{
		ID:     testCategoryID,
		UserID: testUserID, // Categoria pertence ao usuário
	}
	mockCatRepo.On("GetCategoryByID", ctx, testCategoryID).Return(mockCategory, nil)

	// 3. Mock: CreateTransactionInTx
	// Usamos mock.AnythingOfType para não precisar construir a struct exata com timestamps
	mockTxRepo.On("CreateTransactionInTx", ctx, mock.AnythingOfType("*transactions.Transaction")).Return(nil)

	// 4. Mock: UpdateAccountBalanceInTx
	expectedNewBalance := int64(4000) // 5000 - 1000
	mockAccRepo.On("UpdateAccountBalanceInTx", ctx, testAccountID, expectedNewBalance).Return(nil)

	// 5. Execução
	resp, err := service.CreateTransaction(ctx, req, testUserID.String())

	// 6. Assertivas
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(1000), resp.Amount)
	assert.Equal(t, &testCategoryID, resp.CategoryID)
	assert.Equal(t, TypeExpense, resp.Type)

	// Verifica se todos os mocks foram chamados
	mockAccRepo.AssertExpectations(t)
	mockTxRepo.AssertExpectations(t)
	mockCatRepo.AssertExpectations(t)
}

func TestCreateTransaction_Forbidden_Category(t *testing.T) {
	mockAccRepo, _, mockCatRepo, service := setupServiceMocks()
	ctx := context.Background()

	req := CreateTransactionRequest{
		AccountID:  testAccountID.String(),
		Type:       string(TypeExpense),
		Amount:     1000,
		CategoryID: testCategoryID.String(),
	}

	// 1. Mock: GetAccountByIDForUpdate
	mockAccount := &accounts.Account{
		ID:      testAccountID,
		UserID:  testUserID,
		Balance: 5000,
	}
	mockAccRepo.On("GetAccountByIDForUpdate", ctx, testAccountID).Return(mockAccount, nil)

	// 2. Mock: GetCategoryByID (Categoria de OUTRO usuário)
	mockCategory := &categories.Category{
		ID:     testCategoryID,
		UserID: otherUserID, // <-- O PULO DO GATO
	}
	mockCatRepo.On("GetCategoryByID", ctx, testCategoryID).Return(mockCategory, nil)

	// 3. Execução
	resp, err := service.CreateTransaction(ctx, req, testUserID.String())

	// 4. Assertivas
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrForbidden) // Erro de "proibido"

	mockAccRepo.AssertExpectations(t)
	mockCatRepo.AssertExpectations(t)
}

func TestCreateTransaction_CategoryNotFound(t *testing.T) {
	mockAccRepo, _, mockCatRepo, service := setupServiceMocks()
	ctx := context.Background()

	req := CreateTransactionRequest{
		AccountID:  testAccountID.String(),
		Type:       string(TypeExpense),
		Amount:     1000,
		CategoryID: testCategoryID.String(),
	}

	// 1. Mock: GetAccountByIDForUpdate
	mockAccount := &accounts.Account{
		ID:      testAccountID,
		UserID:  testUserID,
		Balance: 5000,
	}
	mockAccRepo.On("GetAccountByIDForUpdate", ctx, testAccountID).Return(mockAccount, nil)

	// 2. Mock: GetCategoryByID (Categoria não existe)
	mockCatRepo.On("GetCategoryByID", ctx, testCategoryID).Return(nil, nil)

	// 3. Execução
	resp, err := service.CreateTransaction(ctx, req, testUserID.String())

	// 4. Assertivas
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrCategoryNotFound)

	mockAccRepo.AssertExpectations(t)
	mockCatRepo.AssertExpectations(t)
}

func TestCreateTransaction_InsufficientFunds(t *testing.T) {
	mockAccRepo, _, mockCatRepo, service := setupServiceMocks()
	ctx := context.Background()

	req := CreateTransactionRequest{
		AccountID:  testAccountID.String(),
		Type:       string(TypeExpense),
		Amount:     10000, // <-- Valor alto
		CategoryID: testCategoryID.String(),
	}

	// 1. Mock: GetAccountByIDForUpdate (Saldo baixo)
	mockAccount := &accounts.Account{
		ID:      testAccountID,
		UserID:  testUserID,
		Balance: 5000, // <-- Saldo baixo
	}
	mockAccRepo.On("GetAccountByIDForUpdate", ctx, testAccountID).Return(mockAccount, nil)

	// 2. Mock: GetCategoryByID (Categoria OK)
	mockCategory := &categories.Category{
		ID:     testCategoryID,
		UserID: testUserID,
	}
	mockCatRepo.On("GetCategoryByID", ctx, testCategoryID).Return(mockCategory, nil)

	// 3. Execução
	resp, err := service.CreateTransaction(ctx, req, testUserID.String())

	// 4. Assertivas
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrInsufficientFunds)

	mockAccRepo.AssertExpectations(t)
	mockCatRepo.AssertExpectations(t)
}
