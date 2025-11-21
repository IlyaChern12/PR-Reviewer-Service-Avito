package main

import (
	"net/http"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/config"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/logger"
)

func main() {
	// подгружаем конфиг
	cfg := config.LoadConfig()

	// логгер
	logger.Init()
	defer logger.Sugar.Sync()
	logger.Sugar.Infof("Starting PR Reviewer Service on port %s", cfg.Port)

	// проверка готовности сервиса
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// запускаем сервер
	addr := ":" + cfg.Port
	logger.Sugar.Infof("Server is listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Sugar.Fatalf("Failed to start server: %v", err)
	}
}