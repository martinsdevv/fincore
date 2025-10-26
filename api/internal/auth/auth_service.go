package auth

import "context"

type Service interface {
	Register(ctx context.Context, req RegisterRequest) error
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
}

type service struct {
	repo      Repository
	jwtSecret string
}

func NewService(repo Repository, jwtSecret string) Service {
	return &service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) error {
	// TODO: 1. Validar input (já feito pelo handler)
	// TODO: 2. Checar se email já existe (chamar repo.GetUserByEmail)
	// TODO: 3. Hash da senha (bcrypt)
	// TODO: 4. Criar struct User
	// TODO: 5. Chamar repo.CreateUser
	return nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// TODO: 1. Validar input
	// TODO: 2. Buscar usuário (repo.GetUserByEmail)
	// TODO: 3. Comparar senha (bcrypt.CompareHashAndPassword)
	// TODO: 4. Gerar JWT
	// TODO: 5. Retornar LoginResponse
	return nil, nil
}
