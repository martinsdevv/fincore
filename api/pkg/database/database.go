package database

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/martinsdevv/fincore/internal/config"
)

var DB *pgxpool.Pool
var Redis *redis.Client

func ConnectDB(cfg *config.Config) error {
	var err error
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)
	DB, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("não foi possível criar a pool de conexões com o DB: %w", err)
	}
	if err = DB.Ping(context.Background()); err != nil {
		DB.Close()
		return fmt.Errorf("não foi possível pingar o DB: %w", err)
	}
	log.Println("Conectado ao banco com sucesso!")
	return nil
}

func ConnectRedis(cfg *config.Config) error {
	Redis = redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	if _, err := Redis.Ping(context.Background()).Result(); err != nil {
		return fmt.Errorf("não foi possível conectar ao Redis: %w", err)
	}
	log.Println("Conectado ao Redis com Sucesso!")
	return nil
}

func CloseConnections() {
	if DB != nil {
		DB.Close()
		log.Println("Conexão fechada!")
	}
	if Redis != nil {
		if err := Redis.Close(); err != nil {
			log.Printf("Erro ao fechar a conexão com o Redis: %s", err)
		} else {
			log.Println("Conexão com o Redis fechada!")
		}
	}
}
