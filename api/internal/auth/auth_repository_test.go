package auth

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	testPool *pgxpool.Pool // Pool de conexão para os testes
	testCtx  = context.Background()
)

// TestMain é uma função especial que roda ANTES de todos os testes neste pacote.
// Vamos usá-la para configurar o banco de dados.
func TestMain(m *testing.M) {
	// Carrega o .env da pasta raiz (estamos em /api, então subimos um nível)
	err := godotenv.Load("../../../.env")
	if err != nil {
		log.Printf("Aviso: Não foi possível carregar o arquivo .env: %v", err)
		log.Println("Tentando continuar com as variáveis de ambiente existentes (ex: CI)...")
	}

	// Ler variáveis de ambiente (agora elas devem existir!)
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Se MESMO ASSIM não encontrar (ex: estamos no CI)
	if dbHost == "" || dbUser == "" {
		log.Println("Variáveis de ambiente do banco de teste (DB_HOST, DB_USER, etc.) não definidas. Pulando testes de repositório.")
		// Precisamos sair para não falhar
		os.Exit(0)
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	testPool, err = pgxpool.New(testCtx, connString)
	if err != nil {
		log.Fatalf("Não foi possível conectar ao banco de dados de teste: %v\n", err)
	}
	defer testPool.Close()

	// A tabela agora é criada pelo 'migrate' no CI.
	// A chamada para createUsersTable() foi removida daqui.

	code := m.Run()

	os.Exit(code)
}

// A função createUsersTable() foi removida.

// Helper para limpar a tabela antes de cada teste
func truncateUsersTable(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(testCtx, "TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("Não foi possível limpar a tabela 'users': %v", err)
	}
}

// --- Testes do Repositório ---

func TestPgxRepository_CreateUser_e_GetUserByEmail(t *testing.T) {
	if testPool == nil {
		t.Skip("Pulando teste: pool de banco de dados não inicializado.")
	}

	// Criar uma instância real do repositório
	repo := NewRepository(testPool)

	// Limpar a tabela antes de começar
	truncateUsersTable(t, testPool)

	mockUser := &User{
		ID:        uuid.New(),
		FirstName: "João",
		LastName:  "Silva",
		Email:     "joao.silva@teste.com",
		Password:  "hash_da_senha_123",
	}

	t.Run("deve criar um usuário com sucesso", func(t *testing.T) {
		err := repo.CreateUser(testCtx, mockUser)
		if err != nil {
			t.Fatalf("esperava nenhum erro ao criar usuário, mas obteve: %v", err)
		}
	})

	t.Run("deve buscar o usuário recém-criado por email", func(t *testing.T) {
		foundUser, err := repo.GetUserByEmail(testCtx, "joao.silva@teste.com")
		if err != nil {
			t.Fatalf("esperava nenhum erro ao buscar usuário, mas obteve: %v", err)
		}
		if foundUser == nil {
			t.Fatal("esperava encontrar um usuário, mas obteve nil")
		}

		// Validar os campos
		if foundUser.ID != mockUser.ID {
			t.Errorf("ID do usuário não bate. esperado=%v, obtido=%v", mockUser.ID, foundUser.ID)
		}
		if foundUser.Email != mockUser.Email {
			t.Errorf("Email do usuário não bate. esperado=%s, obtido=%s", mockUser.Email, foundUser.Email)
		}
		if foundUser.Password != mockUser.Password {
			t.Errorf("Senha do usuário não bate. esperado=%s, obtido=%s", mockUser.Password, foundUser.Password)
		}
	})

	t.Run("deve retornar (nil, nil) para email não encontrado", func(t *testing.T) {
		foundUser, err := repo.GetUserByEmail(testCtx, "email.nao.existe@teste.com")
		if err != nil {
			t.Fatalf("esperava nenhum erro (nil), mas obteve: %v", err)
		}
		if foundUser != nil {
			t.Fatal("esperava (nil) para usuário, mas um usuário foi retornado")
		}
	})

	// TODO: Adicionar teste para violação de email duplicado
	// t.Run("deve retornar erro ao tentar criar email duplicado", ...)
}
