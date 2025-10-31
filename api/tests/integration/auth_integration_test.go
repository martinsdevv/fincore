package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	// Importar os pacotes da sua aplicação
	// (Ajuste os caminhos se o seu go.mod for diferente)
	"github.com/martinsdevv/fincore/internal/auth"
	"github.com/martinsdevv/fincore/internal/config"
)

var (
	testServer *httptest.Server // Nosso servidor de teste
	testPool   *pgxpool.Pool    // Conexão direta ao banco para verificação
	testCtx    = context.Background()
)

// TestMain vai configurar o servidor real
func TestMain(m *testing.M) {
	// 1. Carregar Config
	// (Estamos em /api/tests/integration, subimos 3 níveis)
	if err := godotenv.Load("../../../.env"); err != nil {
		log.Printf("Aviso: Não foi possível carregar o .env para teste de integração: %v", err)
		log.Println("Tentando continuar com as variáveis de ambiente do CI...")
	}

	// Usamos o config.LoadConfig() real, mas as vars vêm do .env que carregamos
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Erro ao carregar config: %v", err)
	}

	// Forçar o uso do banco de dados de teste (garantia)
	// (Garantia de que o config.LoadConfig() pegou as vars de ambiente do CI)
	if os.Getenv("DB_NAME") != "" {
		cfg.DBName = os.Getenv("DB_NAME")
	}

	// 2. Conectar ao Banco de Teste
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	testPool, err = pgxpool.New(testCtx, connString)
	if err != nil {
		log.Fatalf("Não foi possível conectar ao banco de dados de teste: %v", err)
	}
	defer testPool.Close()

	// A tabela agora é criada pelo 'migrate' no CI.
	// A chamada para createUsersTable() foi removida daqui.

	// 3. Montar a Aplicação Real
	repo := auth.NewRepository(testPool)
	service := auth.NewService(repo, cfg.JWTSecret)
	handler := auth.NewHandler(service, cfg.JWTSecret)

	// 4. Configurar o Roteador Real
	r := chi.NewMux()
	handler.RegisterRoutes(r) // Registra as rotas /auth/register e /auth/login

	// 5. Iniciar o Servidor de Teste
	testServer = httptest.NewServer(r)
	defer testServer.Close()

	// Rodar os testes
	code := m.Run()
	os.Exit(code)
}

// --- Testes de Integração ---

func TestAuthIntegration_RegisterAndLogin(t *testing.T) {
	// Limpar o banco antes do teste
	truncateUsersTable(t, testPool)

	// --- 1. Teste de Registro ---
	t.Run("deve registrar um novo usuário com sucesso", func(t *testing.T) {
		payload := map[string]string{
			"first_name": "Usuario",
			"last_name":  "Integrado",
			"email":      "integrado@teste.com",
			"password":   "SenhaForte123",
		}
		jsonBody, _ := json.Marshal(payload)

		// Faz a chamada HTTP POST real para o servidor de teste
		resp, err := http.Post(
			fmt.Sprintf("%s/auth/register", testServer.URL),
			"application/json",
			bytes.NewBuffer(jsonBody),
		)
		if err != nil {
			t.Fatalf("Erro ao fazer requisição de registro: %v", err)
		}
		defer resp.Body.Close()

		// Verificar Status Code
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("esperava status %d, mas obteve %d", http.StatusCreated, resp.StatusCode)
		}

		// Verificar o banco de dados DIRETAMENTE
		var userEmail string
		err = testPool.QueryRow(testCtx, "SELECT email FROM users WHERE email = $1", "integrado@teste.com").Scan(&userEmail)
		if err != nil {
			t.Fatalf("Erro ao verificar usuário no banco: %v", err)
		}
		if userEmail != "integrado@teste.com" {
			t.Errorf("Usuário não foi encontrado no banco após o registro")
		}
	})

	// --- 2. Teste de Login ---
	t.Run("deve logar com o usuário recém-criado", func(t *testing.T) {
		payload := map[string]string{
			"email":    "integrado@teste.com",
			"password": "SenhaForte123",
		}
		jsonBody, _ := json.Marshal(payload)

		resp, err := http.Post(
			fmt.Sprintf("%s/auth/login", testServer.URL),
			"application/json",
			bytes.NewBuffer(jsonBody),
		)
		if err != nil {
			t.Fatalf("Erro ao fazer requisição de login: %v", err)
		}
		defer resp.Body.Close()

		// Verificar Status Code
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("esperava status %d, mas obteve %d", http.StatusOK, resp.StatusCode)
		}

		// Verificar o corpo da resposta (se tem o token)
		var loginResp auth.LoginResponse // Reutiliza a struct do pacote auth
		if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
			t.Fatalf("Erro ao decodificar resposta JSON de login: %v", err)
		}

		if loginResp.AccessToken == "" {
			t.Error("Esperava um access_token, mas veio vazio")
		}
	})

	// TODO: Adicionar teste para email duplicado (deve retornar 409 ou 500)
	// TODO: Adicionar teste para login com senha errada (deve retornar 401 ou 500)
}

// --- Funções Auxiliares ---

// A função createUsersTable() foi removida.

func truncateUsersTable(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(testCtx, "TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("Não foi possível limpar a tabela 'users': %v", err)
	}
}
