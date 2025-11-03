package transactions

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/martinsdevv/fincore/internal/accounts"
	"github.com/rs/zerolog/log"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrTransactionFailed = errors.New("transaction failed")
	ErrAccountNotFound   = errors.New("account not found")
	ErrForbidden         = errors.New("user does not have permission for this account")
)

type Service interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest, userID string) (*TransactionResponse, error)
	ListTransactionsByAccount(ctx context.Context, accountID string, userID string) ([]TransactionResponse, error)
}

type service struct {
	db               *pgxpool.Pool       // O pool de conexão para iniciar transações
	accountsRepo     accounts.Repository // O repositório de contas (para travar e atualizar saldo)
	transactionsRepo Repository          // O repositório de transações (para inserir o registro)
}

func NewService(db *pgxpool.Pool, accountsRepo accounts.Repository, transactionsRepo Repository) Service {
	return &service{
		db:               db,
		accountsRepo:     accountsRepo,
		transactionsRepo: transactionsRepo,
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
		Category:        req.Category,
		TransactionDate: transactionDate,
		CreatedAt:       time.Now().UTC(),
	}

	txFunc := func(tx pgx.Tx) error {
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

		var newBalance int64
		if transactionType == TypeIncome {
			newBalance = account.Balance + req.Amount
		} else {
			if account.Balance < req.Amount {
				return ErrInsufficientFunds
			}
			newBalance = account.Balance - req.Amount
		}

		if err := s.transactionsRepo.CreateTransactionInTx(ctx, tx, newTransaction); err != nil {
			log.Error().Err(err).Msg("Falha ao criar transação no CreateTransactionInTx")
			return ErrTransactionFailed
		}

		if err := s.accountsRepo.UpdateAccountBalanceInTx(ctx, tx, accountID, newBalance); err != nil {
			log.Error().Err(err).Msg("Falha ao atualizar saldo no UpdateAccountBalanceInTx")
			return ErrTransactionFailed
		}

		return nil
	}

	if err := s.executeTx(ctx, txFunc); err != nil {
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

	// --- VERIFICAÇÃO DE SEGURANÇA ---
	// Antes de listar o extrato, verificamos se o usuário é dono da conta.
	// Usamos o GetAccountByID normal, sem lock (FOR UPDATE), pois é só uma leitura.
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

	transactions, err := s.transactionsRepo.ListTransactionsByAccountID(ctx, accountID)
	if err != nil {
		log.Error().Err(err).Str("accountID", accountIDStr).Msg("Falha ao listar transações do repositório")
		return nil, err
	}

	responses := make([]TransactionResponse, len(transactions))
	for i, tr := range transactions {
		responses[i] = *toTransactionResponse(&tr)
	}

	return responses, nil
}

func toTransactionResponse(tr *Transaction) *TransactionResponse {
	return &TransactionResponse{
		ID:              tr.ID,
		AccountID:       tr.AccountID,
		Type:            tr.Type,
		Amount:          tr.Amount,
		Description:     tr.Description,
		Category:        tr.Category,
		TransactionDate: tr.TransactionDate,
		CreatedAt:       tr.CreatedAt,
	}
}
