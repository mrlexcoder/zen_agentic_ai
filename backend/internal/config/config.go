package config

import (
	"os"
)

type Config struct {
	Port            string
	DBPath          string
	BinanceAPIKey   string
	BinanceSecretKey string
	Mode            string
	MaxHistoricalYears int
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		DBPath:            getEnv("DB_PATH", "./trading.db"),
		BinanceAPIKey:     getEnv("BINANCE_API_KEY", "E3Lde33fIAr9FiaIqsCylc2jLYnW15vUhHeVg26IESnI3EF9Um9sVKtu7CfMRSvp"),
		BinanceSecretKey:  getEnv("BINANCE_SECRET_KEY", "WlzwhMTsOfnDvEeSOEx3VanYn1oUbZcUmVAbAmketY8iCqVjr4dqy0WyZYMIb2Qp"),
		Mode:              getEnv("MODE", "development"),
		MaxHistoricalYears: 20,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}