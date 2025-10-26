package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// MockRepository é a simulação da nossa interface Repository
type MockRepository struct {
	CreateUserFunc     func(ctx context.Context, user *User) error
	GetUserByEmailFunc func(ctx context.Context, email string) (*User, error)
}

func (m *MockRepository) CreateUser(ctx context.Context, user *User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return nil
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func TestService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("deve registrar usuário com sucesso", func(t *testing.T) {
		mockRepo := &MockRepository{
			// Simula que o usuário NÃO existe
			GetUserByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil // nil, nil == usuário não encontrado
			},
			// Simula que a criação do usuário foi bem-sucedida
			CreateUserFunc: func(ctx context.Context, user *User) error {
				return nil
			},
		}

		service := NewService(mockRepo, "test_secret")
		req := RegisterRequest{
			FirstName: "Teste",
			LastName:  "Usuario",
			Email:     "novo@exemplo.com",
			Password:  "senha123",
		}

		err := service.Register(ctx, req)
		if err != nil {
			t.Errorf("esperava nenhum erro, mas obteve %v", err)
		}
	})

	t.Run("deve retornar erro se o email já existe", func(t *testing.T) {
		mockRepo := &MockRepository{
			// Simula que o usuário JÁ existe
			GetUserByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return &User{Email: "existente@exemplo.com"}, nil
			},
		}

		service := NewService(mockRepo, "test_secret")
		req := RegisterRequest{Email: "existente@exemplo.com", Password: "senha123"}

		err := service.Register(ctx, req)
		if !errors.Is(err, ErrEmailConflict) {
			t.Errorf("esperava o erro %v, mas obteve %v", ErrEmailConflict, err)
		}
	})

	t.Run("deve retornar erro em falha do repositório", func(t *testing.T) {
		dbError := errors.New("falha na conexão com o banco")
		mockRepo := &MockRepository{
			// Simula um erro de banco
			GetUserByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, dbError
			},
		}

		service := NewService(mockRepo, "test_secret")
		req := RegisterRequest{Email: "teste@exemplo.com", Password: "senha123"}

		err := service.Register(ctx, req)
		if !errors.Is(err, dbError) {
			t.Errorf("esperava o erro %v, mas obteve %v", dbError, err)
		}
	})
}

func TestService_Login(t *testing.T) {
	ctx := context.Background()
	jwtSecret := "meu_segredo_de_teste"

	// Usuário mockado com senha "senha123" hasheada
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("senha123"), bcrypt.DefaultCost)
	mockUser := &User{
		ID:       uuid.New(),
		Email:    "usuario@exemplo.com",
		Password: string(hashedPassword),
	}

	t.Run("deve logar com sucesso e retornar jwt", func(t *testing.T) {
		mockRepo := &MockRepository{
			GetUserByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				if email == mockUser.Email {
					return mockUser, nil
				}
				return nil, nil // Não encontrado
			},
		}

		service := NewService(mockRepo, jwtSecret)
		req := LoginRequest{Email: "usuario@exemplo.com", Password: "senha123"}

		resp, err := service.Login(ctx, req)
		if err != nil {
			t.Errorf("esperava nenhum erro, mas obteve %v", err)
		}
		if resp == nil || resp.AccessToken == "" {
			t.Fatal("esperava um token de acesso, mas obteve nulo ou vazio")
		}

		// Validar o token (opcional, mas bom)
		token, err := jwt.Parse(resp.AccessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil {
			t.Errorf("falha ao analisar o token gerado: %v", err)
		}
		if !token.Valid {
			t.Error("o token gerado não é válido")
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		if claims["sub"] != mockUser.ID.String() {
			t.Error("a claim 'sub' do token está incorreta")
		}
	})

	t.Run("deve retornar erro se o usuário não for encontrado", func(t *testing.T) {
		mockRepo := &MockRepository{
			GetUserByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil // Usuário não encontrado
			},
		}

		service := NewService(mockRepo, jwtSecret)
		req := LoginRequest{Email: "naoencontrado@exemplo.com", Password: "senha123"}

		resp, err := service.Login(ctx, req)
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("esperava o erro %v, mas obteve %v", ErrInvalidCredentials, err)
		}
		if resp != nil {
			t.Error("esperava resposta nula em caso de falha")
		}
	})

	t.Run("deve retornar erro em caso de senha incorreta", func(t *testing.T) {
		mockRepo := &MockRepository{
			GetUserByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return mockUser, nil // Retorna o usuário mockado
			},
		}

		service := NewService(mockRepo, jwtSecret)
		req := LoginRequest{Email: "usuario@exemplo.com", Password: "SENHA_ERRADA"}

		resp, err := service.Login(ctx, req)
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("esperava o erro %v, mas obteve %v", ErrInvalidCredentials, err)
		}
		if resp != nil {
			t.Error("esperava resposta nula em caso de falha")
		}
	})
}
