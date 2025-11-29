package integration

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/handler"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/router"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/service"
	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

var baseURL string

// main для тестов
func TestMain(m *testing.M) {
	// подключение к тестовой БД
	dbConn := InitTestDB()

	// создаём мок-логгер
	mockLogger := zap.NewNop().Sugar()
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	// gin.DefaultErrorWriter = io.Discard

	// репозитории
	userRepo := repository.NewUserRepo(dbConn, mockLogger)
	teamRepo := repository.NewTeamRepo(dbConn, mockLogger)
	prRepo := repository.NewPullRequestRepo(dbConn)

	// сервисы
	userService := service.NewUserService(userRepo, mockLogger)
	teamService := service.NewTeamService(teamRepo, dbConn, mockLogger)
	prService := service.NewPullRequestService(prRepo, userService, mockLogger)

	// хэндлеры
	userHandler := handler.NewUserHandler(userService, mockLogger)
	teamHandler := handler.NewTeamHandler(teamService, prService, userService, mockLogger)
	prHandler := handler.NewPullRequestHandler(prService, mockLogger)
	statsHandler := handler.NewStatsHandler(prService, userService, teamService, mockLogger)

	// роутер
	router := router.NewRouter(userHandler, teamHandler, prHandler, statsHandler)

	// сервер
	addr := ":8081"
	baseURL = "http://localhost" + addr
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("failed to start test server: %v\n", err)
		}
	}()

	// запуск тестов
	code := m.Run()

	// остановка сервера
	_ = srv.Close()
	os.Exit(code)
}