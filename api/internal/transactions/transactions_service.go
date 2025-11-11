package transactions

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/martinsdevv/fincore/internal/accounts"
	"github.com/martinsdevv/fincore/internal/categories" // <-- IMPORT NOVO
	"github.com/rs/zerolog/log"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrTransactionFailed = errors.New("transaction failed")
	ErrAccountNotFound   = errors.New("account not found")
	ErrForbidden         = errors.New("user does not have permission for this account")
	ErrCategoryNotFound  = errors.New("category not found") // <-- ERRO NOVO
)

type Service interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest, userID string) (*TransactionResponse, error)
	ListTransactionsByAccount(ctx context.Context, accountID string, userID string) ([]TransactionResponse, error)
}

type service struct {
	db               *pgxpool.Pool
	accountsRepo     accounts.Repository
	transactionsRepo Repository
	categoriesRepo   categories.Repository // <-- DEPENDÊNCIA NOVA
}

// Assinatura do NewService mudou
func NewService(db *pgxpool.Pool, accountsRepo accounts.Repository, transactionsRepo Repository, categoriesRepo categories.Repository) Service {
	return &service{
		db:               db,
		accountsRepo:     accountsRepo,
		transactionsRepo: transactionsRepo,
		categoriesRepo:   categoriesRepo, // <-- DEPENDÊNCIA NOVA
	}
}

func (s *service) executeTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *service) CreateTransaction(ctx context.Context, req CreateTransactionRequest, userIDStr string) (*TransactionResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, errors.New("invalid account ID")
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, errors.New("invalid category ID")
	}

	transactionDate, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		transactionDate = time.Now().UTC()
	}

	transactionType := TransactionType(req.Type)

	newTransaction := &Transaction{
		ID:              uuid.New(),
		AccountID:       accountID,
		Type:            transactionType,
		Amount:          req.Amount,
		Description:     req.Description,
		CategoryID:      &categoryID,
		TransactionDate: transactionDate,
		CreatedAt:       time.Now().UTC(),
	}

	// --- MUDANÇA AQUI ---
	// Declaramos a categoria aqui fora para podermos acessar o nome dela
	// depois que a transação (txFunc) for executada.
	var category *categories.Category
	// --- FIM DA MUDANÇA ---

	txFunc := func(tx pgx.Tx) error {
		// 1. Validar a Conta
		account, err := s.accountsRepo.GetAccountByIDForUpdate(ctx, tx, accountID)
		if err != nil {
			log.Error().Err(err).Msg("Falha ao obter conta no GetAccountByIDForUpdate")
			return ErrTransactionFailed
		}
		if account == nil {
			return ErrAccountNotFound
		}
		if account.UserID != userID {
			return ErrForbidden
		}

		// 2. Validar a Categoria
		// --- MUDANÇA AQUI ---
		// Removemos o 'var' para atribuir à variável externa
		category, err = s.categoriesRepo.GetCategoryByID(ctx, categoryID)
		// --- FIM DA MUDANÇA ---
		if err != nil {
			log.Error().Err(err).Msg("Falha ao obter categoria no GetCategoryByID")
			return ErrTransactionFailed
		}
		if category == nil {
			return ErrCategoryNotFound
		}
		if category.UserID != userID {
			log.Warn().Str("userID", userID.String()).Str("categoryOwnerID", category.UserID.String()).Msg("Forbidden category usage attempt")
			return ErrForbidden
		}

		// 3. Validar Saldo
		var newBalance int64
		if transactionType == TypeIncome {
			newBalance = account.Balance + req.Amount
		} else {
			if account.Balance < req.Amount {
				return ErrInsufficientFunds
			}
			newBalance = account.Balance - req.Amount
		}

		// 4. Criar Transação
		if err := s.transactionsRepo.CreateTransactionInTx(ctx, tx, newTransaction); err != nil {
			log.Error().Err(err).Msg("Falha ao criar transação no CreateTransactionInTx")
			return ErrTransactionFailed
		}

		// 5. Atualizar Saldo
		if err := s.accountsRepo.UpdateAccountBalanceInTx(ctx, tx, accountID, newBalance); err != nil {
			log.Error().Err(err).Msg("Falha ao atualizar saldo no UpdateAccountBalanceInTx")
			return ErrTransactionFailed
		}

		return nil
	}

	if err := s.executeTx(ctx, txFunc); err != nil {
		return nil, err
	}

	// --- MUDANÇA AQUI ---
	// Não podemos mais usar 'toTransactionResponse(newTransaction)'
	// porque 'newTransaction' é do tipo *Transaction, e o helper espera *ListTransactionsRow.
	// Em vez disso, construímos a resposta manualmente,
	// já que temos 'newTransaction' e a 'category' que buscamos.
	var categoryName *string
	if category != nil {
		categoryName = &category.Name
	}

	return &TransactionResponse{
		ID:              newTransaction.ID,
		AccountID:       newTransaction.AccountID,
		Type:            newTransaction.Type,
		Amount:          newTransaction.Amount,
		Description:     newTransaction.Description,
		CategoryID:      newTransaction.CategoryID,
		CategoryName:    categoryName, // <-- Retornamos o nome
		TransactionDate: newTransaction.TransactionDate,
		CreatedAt:       newTransaction.CreatedAt,
	}, nil
	// --- FIM DA MUDANÇA ---
}

func (s *service) ListTransactionsByAccount(ctx context.Context, accountIDStr string, userIDStr string) ([]TransactionResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return nil, errors.New("invalid account ID")
	}

	// Checagem de segurança (continua igual)
	account, err := s.accountsRepo.GetAccountByID(ctx, accountID)
	if err != nil {
		log.Error().Err(err).Str("accountID", accountIDStr).Msg("Falha ao checar dono da conta")
		return nil, err
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}
	if account.UserID != userID {
		log.Warn().Str("userID", userIDStr).Str("accountOwnerID", account.UserID.String()).Msg("Tentativa de listar extrato de conta alheia")
		return nil, ErrForbidden
	}

	// --- MUDANÇA AQUI ---
	// O repo agora retorna 'ListTransactionsRow'
	transactions, err := s.transactionsRepo.ListTransactionsByAccountID(ctx, accountID)
	if err != nil {
		log.Error().Err(err).Str("accountID", accountIDStr).Msg("Falha ao listar transações do repositório")
		return nil, err
	}

	// Mapeia de 'ListTransactionsRow' para 'TransactionResponse'
	responses := make([]TransactionResponse, len(transactions))
	for i, tr := range transactions {
		responses[i] = *toTransactionResponse(&tr) // <-- Usa o helper atualizado
	}

	return responses, nil
}

// Helper de conversão
func toTransactionResponse(tr *ListTransactionsRow) *TransactionResponse {
	resp := &TransactionResponse{
		ID:              tr.ID,
		AccountID:       tr.AccountID,
		Type:            tr.Type,
		Amount:          tr.Amount,
		Description:     tr.Description,
		CategoryID:      tr.CategoryID,
		TransactionDate: tr.TransactionDate,
		CreatedAt:       tr.CreatedAt,
	}

	// Converte sql.NullString (do DB) para *string (do JSON)
	if tr.CategoryName.Valid {
		resp.CategoryName = &tr.CategoryName.String
	}

	return resp
}
