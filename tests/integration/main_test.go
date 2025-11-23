package integration

import (
	"log"
	"os"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	// загружаем .env
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found")
	}

	// инициализация тестовой базы и миграций
	InitTestDB()

	// запускаем тесты
	os.Exit(m.Run())
}
