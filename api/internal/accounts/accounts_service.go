package accounts

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrForbidden       = errors.New("user does not have permission for this account")
)

type Service interface {
	CreateAccount(ctx context.Context, req CreateAccountRequest, userID string) (*AccountResponse, error)
	GetAccount(ctx context.Context, accountID string, userID string) (*AccountResponse, error)
	ListAccounts(ctx context.Context, userID string) ([]AccountResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) parseAndValidateIDs(userIDStr string, accountIDStr ...string) (uuid.UUID, uuid.UUID, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Warn().Err(err).Msg("Invalid User UUID format")
		return uuid.Nil, uuid.Nil, errors.New("invalid user ID")
	}

	var accountID uuid.UUID
	if len(accountIDStr) > 0 && accountIDStr[0] != "" {
		accountID, err = uuid.Parse(accountIDStr[0])
		if err != nil {
			log.Warn().Err(err).Msg("Invalid Account UUID format")
			return uuid.Nil, uuid.Nil, errors.New("invalid account ID")
		}
	}

	return userID, accountID, nil
}

func (s *service) CreateAccount(ctx context.Context, req CreateAccountRequest, userIDStr string) (*AccountResponse, error) {
	userID, _, err := s.parseAndValidateIDs(userIDStr)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	account := &Account{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      req.Name,
		Type:      req.Type,
		Balance:   req.InitialBalance,
		Currency:  req.Currency,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateAccount(ctx, account); err != nil {
		log.Error().Err(err).Msg("Failed to create account in repository")
		return nil, err
	}

	return toAccountResponse(account), nil
}

func (s *service) GetAccount(ctx context.Context, accountIDStr string, userIDStr string) (*AccountResponse, error) {
	userID, accountID, err := s.parseAndValidateIDs(userIDStr, accountIDStr)
	if err != nil {
		return nil, err
	}

	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		log.Error().Err(err).Str("accountID", accountIDStr).Msg("Failed to get account from repository")
		return nil, err
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}

	// Verifica se a conta pertence ao usuário que fez a requisição
	if account.UserID != userID {
		log.Warn().Str("userID", userIDStr).Str("accountOwnerID", account.UserID.String()).Msg("Forbidden access attempt")
		return nil, ErrForbidden
	}

	return toAccountResponse(account), nil
}

func (s *service) ListAccounts(ctx context.Context, userIDStr string) ([]AccountResponse, error) {
	userID, _, err := s.parseAndValidateIDs(userIDStr)
	if err != nil {
		return nil, err
	}

	accounts, err := s.repo.ListAccountsByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("userID", userIDStr).Msg("Failed to list accounts from repository")
		return nil, err
	}

	responses := make([]AccountResponse, len(accounts))
	for i, acc := range accounts {
		responses[i] = *toAccountResponse(&acc)
	}

	return responses, nil
}

func toAccountResponse(acc *Account) *AccountResponse {
	return &AccountResponse{
		ID:        acc.ID,
		UserID:    acc.UserID,
		Name:      acc.Name,
		Type:      acc.Type,
		Balance:   acc.Balance,
		Currency:  acc.Currency,
		CreatedAt: acc.CreatedAt,
	}
}
