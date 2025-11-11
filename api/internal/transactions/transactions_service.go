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

	// --- MUDANÇA AQUI ---
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, errors.New("invalid category ID")
	}

	transactionDate, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		transactionDate = time.Now().UTC()
	}

	transactionType := TransactionType(req.Type)

	// --- MUDANÇA AQUI ---
	newTransaction := &Transaction{
		ID:              uuid.New(),
		AccountID:       accountID,
		Type:            transactionType,
		Amount:          req.Amount,
		Description:     req.Description,
		CategoryID:      &categoryID, // <-- MUDOU AQUI
		TransactionDate: transactionDate,
		CreatedAt:       time.Now().UTC(),
	}

	txFunc := func(tx pgx.Tx) error {
		// 1. Validar a Conta (igual a antes)
		account, err := s.accountsRepo.GetAccountByIDForUpdate(ctx, tx, accountID)
		if err != nil {
			log.Error().Err(err).Msg("Falha ao obter conta no GetAccountByIDForUpdate")
			return ErrTransactionFailed
		}
		if account == nil {
			return ErrAccountNotFound
		}
		if account.UserID != userID {
			return ErrForbidden // Usuário tentando usar conta de outro
		}

		// --- LÓGICA NOVA: Validar a Categoria ---
		// (Não usamos "ForUpdate" aqui, só precisamos ler)
		category, err := s.categoriesRepo.GetCategoryByID(ctx, categoryID)
		if err != nil {
			log.Error().Err(err).Msg("Falha ao obter categoria no GetCategoryByID")
			return ErrTransactionFailed
		}
		if category == nil {
			return ErrCategoryNotFound
		}
		if category.UserID != userID {
			log.Warn().Str("userID", userID.String()).Str("categoryOwnerID", category.UserID.String()).Msg("Forbidden category usage attempt")
			return ErrForbidden // Usuário tentando usar categoria de outro
		}
		// --- FIM DA LÓGICA NOVA ---

		// 3. Validar Saldo (igual a antes)
		var newBalance int64
		if transactionType == TypeIncome {
			newBalance = account.Balance + req.Amount
		} else {
			if account.Balance < req.Amount {
				return ErrInsufficientFunds
			}
			newBalance = account.Balance - req.Amount
		}

		// 4. Criar Transação (igual a antes, mas repo já está atualizado)
		if err := s.transactionsRepo.CreateTransactionInTx(ctx, tx, newTransaction); err != nil {
			log.Error().Err(err).Msg("Falha ao criar transação no CreateTransactionInTx")
			return ErrTransactionFailed
		}

		// 5. Atualizar Saldo (igual a antes)
		if err := s.accountsRepo.UpdateAccountBalanceInTx(ctx, tx, accountID, newBalance); err != nil {
			log.Error().Err(err).Msg("Falha ao atualizar saldo no UpdateAccountBalanceInTx")
			return ErrTransactionFailed
		}

		return nil
	}

	if err := s.executeTx(ctx, txFunc); err != nil {
		// O handler.go vai mapear esses erros para 403, 404, 422...
		return nil, err
	}

	return toTransactionResponse(newTransaction), nil
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

	// Nenhuma mudança de lógica aqui.
	// A checagem de segurança é a mesma.
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

	// O repo já foi atualizado, então essa chamada funciona.
	transactions, err := s.transactionsRepo.ListTransactionsByAccountID(ctx, accountID)
	if err != nil {
		log.Error().Err(err).Str("accountID", accountIDStr).Msg("Falha ao listar transações do repositório")
		return nil, err
	}

	responses := make([]TransactionResponse, len(transactions))
	for i, tr := range transactions {
		// toTransactionResponse foi atualizado
		responses[i] = *toTransactionResponse(&tr)
	}

	return responses, nil
}

// Helper de conversão
func toTransactionResponse(tr *Transaction) *TransactionResponse {
	return &TransactionResponse{
		ID:              tr.ID,
		AccountID:       tr.AccountID,
		Type:            tr.Type,
		Amount:          tr.Amount,
		Description:     tr.Description,
		CategoryID:      tr.CategoryID, // <-- MUDOU AQUI
		TransactionDate: tr.TransactionDate,
		CreatedAt:       tr.CreatedAt,
	}
}
