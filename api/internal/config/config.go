package config

import (
	"github.com/spf13/viper"
)

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
	v := viper.New()
	v.AutomaticEnv()

	// Garante que as variáveis de ambiente sejam consideradas no Unmarshal
	for _, k := range []string{
		"API_PORT",
		"JWT_SECRET",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"REDIS_ADDR",
	} {
		_ = v.BindEnv(k)
	}

	v.SetDefault("API_PORT", "8080")
	v.SetDefault("DB_PORT", "5432")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
