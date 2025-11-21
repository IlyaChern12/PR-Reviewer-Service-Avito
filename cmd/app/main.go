package main

import (
	"net/http"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/config"
	"go.uber.org/zap"
)

func main() {
	// подгружаем конфиг
	cfg := config.LoadConfig()

	// логгер
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	sugar.Infof("Starting PR Reviewer Service on port %s", cfg.Port)

	// проверка готовности сервиса
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// запускаем сервер
	addr := ":" + cfg.Port
	sugar.Infof("Server is listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		sugar.Fatalf("Failed to start server: %v", err)
	}
}