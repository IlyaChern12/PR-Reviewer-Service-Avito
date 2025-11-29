package main

import (
	"fmt"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/config"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/handler"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/logger"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/db"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/router"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/service"
)

func main() {
	// конфиг
	cfg := config.LoadConfig()

	// логгер
	logger.Init()
	defer func() {
		if err := logger.Sugar.Sync(); err != nil {
			logger.Sugar.Warnf("logger sync failed: %v", err)
		}
	}()

	// подключение к БД
	dbConn, err := db.NewPostgresDB(cfg)
	if err != nil {
		logger.Sugar.Fatalf("failed to connect to DB: %v", err)
	}

	// репо
	userRepo := repository.NewUserRepo(dbConn, logger.Sugar)
	teamRepo := repository.NewTeamRepo(dbConn, logger.Sugar)
	prRepo := repository.NewPullRequestRepo(dbConn)

	// сервисы
	userService := service.NewUserService(userRepo, logger.Sugar)
	teamService := service.NewTeamService(teamRepo, dbConn, logger.Sugar)
	prService := service.NewPullRequestService(prRepo, userService, logger.Sugar)

	// хэндлеры
	userHandler := handler.NewUserHandler(userService, logger.Sugar)
	teamHandler := handler.NewTeamHandler(teamService, prService, userService, logger.Sugar)
	prHandler := handler.NewPullRequestHandler(prService, logger.Sugar)
	statsHandler := handler.NewStatsHandler(prService, userService, teamService, logger.Sugar)

	// роутер
	router := router.NewRouter(userHandler, teamHandler, prHandler, statsHandler)

	// запуск сервера
	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Sugar.Infof("server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		logger.Sugar.Fatalf("failed to start server: %v", err)
	}
}