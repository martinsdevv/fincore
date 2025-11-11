package categories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Erros de negócio específicos deste módulo
var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrForbidden        = errors.New("user does not have permission for this category")
	ErrInvalidUUID      = errors.New("invalid UUID format")
)

type Service interface {
	CreateCategory(ctx context.Context, req CreateCategoryRequest, userID string) (*CategoryResponse, error)
	GetCategory(ctx context.Context, categoryID string, userID string) (*CategoryResponse, error)
	ListCategories(ctx context.Context, userID string) ([]CategoryResponse, error)
	UpdateCategory(ctx context.Context, categoryID string, req UpdateCategoryRequest, userID string) (*CategoryResponse, error)
	DeleteCategory(ctx context.Context, categoryID string, userID string) error
}

// service (Implementação)
type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Helper para não repetir a validação de UUIDs
// (Baseado no que você fez em accounts/service.go)
func (s *service) parseAndValidateIDs(userIDStr string, categoryIDStr string) (uuid.UUID, uuid.UUID, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Warn().Err(err).Msg("Invalid User UUID format")
		return uuid.Nil, uuid.Nil, ErrInvalidUUID
	}

	var categoryID uuid.UUID
	if categoryIDStr != "" {
		categoryID, err = uuid.Parse(categoryIDStr)
		if err != nil {
			log.Warn().Err(err).Msg("Invalid Category UUID format")
			return uuid.Nil, uuid.Nil, ErrInvalidUUID
		}
	}

	return userID, categoryID, nil
}

func (s *service) CreateCategory(ctx context.Context, req CreateCategoryRequest, userIDStr string) (*CategoryResponse, error) {
	userID, _, err := s.parseAndValidateIDs(userIDStr, "")
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	cat := &Category{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateCategory(ctx, cat); err != nil {
		// TODO: Mapear erro de constraint 'uq_user_category_name' para um 'ErrCategoryNameConflict'
		log.Error().Err(err).Msg("Failed to create category in repository")
		return nil, err
	}

	return toCategoryResponse(cat), nil
}

func (s *service) GetCategory(ctx context.Context, categoryIDStr string, userIDStr string) (*CategoryResponse, error) {
	userID, categoryID, err := s.parseAndValidateIDs(userIDStr, categoryIDStr)
	if err != nil {
		return nil, err
	}

	cat, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		log.Error().Err(err).Str("categoryID", categoryIDStr).Msg("Failed to get category from repository")
		return nil, err
	}
	if cat == nil {
		return nil, ErrCategoryNotFound
	}

	// Checagem de segurança
	if cat.UserID != userID {
		log.Warn().Str("userID", userIDStr).Str("categoryOwnerID", cat.UserID.String()).Msg("Forbidden access attempt")
		return nil, ErrForbidden
	}

	return toCategoryResponse(cat), nil
}

func (s *service) ListCategories(ctx context.Context, userIDStr string) ([]CategoryResponse, error) {
	userID, _, err := s.parseAndValidateIDs(userIDStr, "")
	if err != nil {
		return nil, err
	}

	categories, err := s.repo.ListCategoriesByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("userID", userIDStr).Msg("Failed to list categories from repository")
		return nil, err
	}

	responses := make([]CategoryResponse, len(categories))
	for i, cat := range categories {
		// Cuidado para não pegar o ponteiro da variável 'cat' do loop
		tempCat := cat
		responses[i] = *toCategoryResponse(&tempCat)
	}

	return responses, nil
}

func (s *service) UpdateCategory(ctx context.Context, categoryIDStr string, req UpdateCategoryRequest, userIDStr string) (*CategoryResponse, error) {
	userID, categoryID, err := s.parseAndValidateIDs(userIDStr, categoryIDStr)
	if err != nil {
		return nil, err
	}

	// 1. Pega a categoria (e o lock de permissão)
	cat, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		log.Error().Err(err).Str("categoryID", categoryIDStr).Msg("Failed to get category for update")
		return nil, err
	}
	if cat == nil {
		return nil, ErrCategoryNotFound
	}

	// 2. Checa a permissão
	if cat.UserID != userID {
		log.Warn().Str("userID", userIDStr).Str("categoryOwnerID", cat.UserID.String()).Msg("Forbidden update attempt")
		return nil, ErrForbidden
	}

	// 3. Atualiza os campos
	cat.Name = req.Name
	cat.UpdatedAt = time.Now().UTC()

	// 4. Salva no banco
	if err := s.repo.UpdateCategory(ctx, cat); err != nil {
		// TODO: Mapear erro de constraint 'uq_user_category_name'
		log.Error().Err(err).Str("categoryID", categoryIDStr).Msg("Failed to update category in repository")
		return nil, err
	}

	return toCategoryResponse(cat), nil
}

func (s *service) DeleteCategory(ctx context.Context, categoryIDStr string, userIDStr string) error {
	userID, categoryID, err := s.parseAndValidateIDs(userIDStr, categoryIDStr)
	if err != nil {
		return err
	}

	// 1. Pega a categoria (e o lock de permissão)
	cat, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		log.Error().Err(err).Str("categoryID", categoryIDStr).Msg("Failed to get category for delete")
		return err
	}
	if cat == nil {
		return ErrCategoryNotFound
	}

	// 2. Checa a permissão
	if cat.UserID != userID {
		log.Warn().Str("userID", userIDStr).Str("categoryOwnerID", cat.UserID.String()).Msg("Forbidden delete attempt")
		return ErrForbidden
	}

	// 3. Deleta do banco
	if err := s.repo.DeleteCategory(ctx, categoryID); err != nil {
		log.Error().Err(err).Str("categoryID", categoryIDStr).Msg("Failed to delete category in repository")
		return err
	}

	return nil
}
