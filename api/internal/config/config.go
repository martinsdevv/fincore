package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config armazena todas as configurações da aplicação.
// As tags 'mapstructure' dizem ao viper qual variável de ambiente
// corresponde a qual campo da struct.

type Config struct {
	APIPort    string `mapstructure:"API_PORT"`
	JWTSecret  string `mapstructure:"JWT_SECRET"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	RedisAddr  string `mapstructure:"REDIS_ADDR"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("../..")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("erro ao ler .env: %w", err)
		}
	}

	// viper.BindEnv("API_PORT")
	// viper.BindEnv("JWT_SECRET")
	// viper.BindEnv("DB_HOST")
	// viper.BindEnv("DB_PORT")
	// viper.BindEnv("DB_USER")
	// viper.BindEnv("DB_PASSWORD")
	// viper.BindEnv("DB_NAME")
	// viper.BindEnv("REDIS_ADDR")

	viper.SetDefault("DB_PORT", "5432")

	viper.AutomaticEnv()

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Validação mínima para evitar conexões com credenciais vazias
	if cfg.DBUser == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("variáveis de DB ausentes: verifique DB_USER e DB_NAME no .env ou ambiente")
	}

	return &cfg, nil
}
