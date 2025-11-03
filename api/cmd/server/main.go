package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/martinsdevv/fincore/internal/accounts"
	"github.com/martinsdevv/fincore/internal/auth"
	"github.com/martinsdevv/fincore/internal/config"
	"github.com/martinsdevv/fincore/pkg/database"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Info().Msg("Logger Iniciado")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Não foi possível carregar as configurações")
	}

	if err := runMigrations(cfg); err != nil {
		log.Fatal().Err(err).Msg("Falha ao rodar migrações")
	}
	log.Info().Msg("Migrações do banco de dados aplicadas com sucesso.")

	if err := database.ConnectDB(cfg); err != nil {
		log.Fatal().Err(err).Msg("Não foi possível conectar ao banco")
	}
	if err := database.ConnectRedis(cfg); err != nil {
		log.Fatal().Err(err).Msg("Não foi possível conectar ao redis")
	}

	defer database.CloseConnections()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.DB.Ping(context.Background()); err != nil {
			log.Error().Err(err).Msg("Falha no health check (DB Ping)")
			http.Error(w, "DB Error", http.StatusInternalServerError)
			return
		}

		if _, err := database.Redis.Ping(context.Background()).Result(); err != nil {
			log.Error().Err(err).Msg("Falha no health check (Redis Ping)")
			http.Error(w, "Redis Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status": "ok", "db": "connected", "cache": "connected"}`))
		if err != nil {
			log.Error().Err(err).Msg("Erro ao escrever resposta do health check")
		}
	})

	authRepo := auth.NewRepository(database.DB)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authSvc, cfg.JWTSecret)

	accountsRepo := accounts.NewRepository(database.DB)
	accountsSvc := accounts.NewService(accountsRepo)
	accountsHandler := accounts.NewHandler(accountsSvc)

	// --- Rotas Públicas ---
	authHandler.RegisterRoutes(r)

	// --- Rotas Protegidas ---
	r.Group(func(r chi.Router) {
		// Aplica o middleware de autenticação a este grupo
		r.Use(authHandler.AuthMiddleware)

		// Rota de teste /me
		r.Get("/auth/me", authHandler.GetMe)

		// Rotas do módulo accounts
		accountsHandler.RegisterRoutes(r)
	})

	serverAddr := fmt.Sprintf(":%s", cfg.APIPort)
	server := &http.Server{Addr: serverAddr, Handler: r}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Msgf("Servidor Fincore API iniciado na porta %s", cfg.APIPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Não foi possível iniciar o servidor")
		}
	}()

	<-stop

	log.Info().Msg("Servidor recebendo sinal de parada. Desligando...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Falha no shutdown gracioso do servidor")
	}

	log.Info().Msg("Servidor desligado com sucesso.")
}

func runMigrations(cfg *config.Config) error {
	dsn := fmt.Sprintf("pgx5://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	m, err := migrate.New("file://./migrations", dsn)
	if err != nil {
		return fmt.Errorf("falha ao instanciar migrate: %w", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info().Msg("Nenhuma nova migração para aplicar.")
			return nil
		}
		return fmt.Errorf("falha ao aplicar migrações 'up': %w", err)
	}

	return nil
}
