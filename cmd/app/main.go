package main

import (
	"fmt"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/config"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/handler"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/logger"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/repository/db"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// Подгружаем конфиг
	cfg := config.LoadConfig()

	// Логгер
	logger.Init()
	defer func() {
		if err := logger.Sugar.Sync(); err != nil {
			logger.Sugar.Warnf("logger sync failed: %v", err)
		}
	}()

	logger.Sugar.Infof("Starting PR Reviewer Service on port %s", cfg.Port)

	// Подключение к БД
	dbConn, err := db.NewPostgresDB(cfg)
	if err != nil {
		logger.Sugar.Fatalf("failed to connect to DB: %v", err)
	}
	prRepo := repository.NewPullRequestRepo(dbConn)
	userRepo := repository.NewUserRepo(dbConn, logger.Sugar)
	teamRepo := repository.NewTeamRepo(dbConn, logger.Sugar)

	// Сервисы
	userService := service.NewUserService(userRepo, logger.Sugar)
	teamService := service.NewTeamService(teamRepo, dbConn, logger.Sugar)
	prService := service.NewPullRequestService(prRepo, userService, logger.Sugar)

	// Хэндлеры
	userHandler := handler.NewUserHandler(userService, logger.Sugar)
	teamHandler := handler.NewTeamHandler(teamService, logger.Sugar)
	prHandler := handler.NewPullRequestHandler(prService, logger.Sugar)
	statsHandler := handler.NewStatsHandler(prService, userService, teamService, logger.Sugar)

	// Gin роутер
	router := gin.Default()

	// Health-check
	router.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// Роуты API
	setupRoutes(router, userHandler, teamHandler, prHandler, statsHandler)

	// Запуск сервера
	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Sugar.Infof("server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		logger.Sugar.Fatalf("failed to start server: %v", err)
	}
}

func setupRoutes(router *gin.Engine, userH *handler.UserHandler, teamH *handler.TeamHandler, prH *handler.PullRequestHandler, stats *handler.StatsHandler) {
	// пользователи
	router.POST("/users/setIsActive", userH.SetIsActive)
	router.GET("/users/getReview", userH.GetReviewPR)

	// команды
	router.POST("/team/add", teamH.CreateTeam)
	router.GET("/team/get", teamH.GetTeam)

	// пулл реквесты
	router.POST("/pullRequest/create", prH.CreatePR)
	router.POST("/pullRequest/merge", prH.MergePR)
	router.POST("/pullRequest/reassign", prH.ReassignReviewer)

	// статистика
	router.GET("/stats", stats.GetStats)
}
