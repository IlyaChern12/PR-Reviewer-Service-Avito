package config

import "os"

type Config struct {
	Port       string
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
}

// получение переменной с возвратом дефолтного значения
func getEnv(key, dflt string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return dflt
}

// подгрузка конфигов
func LoadConfig() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "db"),
		DBUser:     getEnv("DB_USER", "pruser"),
		DBPassword: getEnv("DB_PASSWORD", "prpass"),
		DBName:     getEnv("DB_NAME", "prdb"),
	}
}
