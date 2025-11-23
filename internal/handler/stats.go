package handler

import (
	"net/http"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StatsHandler struct {
	prService   *service.PullRequestService
	userService *service.UserService
	teamService *service.TeamService
	logger      *zap.SugaredLogger
}

func NewStatsHandler(pr *service.PullRequestService, user *service.UserService, team *service.TeamService, logger *zap.SugaredLogger) *StatsHandler {
	return &StatsHandler{
		prService:   pr,
		userService: user,
		teamService: team,
		logger:      logger,
	}
}

// GET /stats
func (h *StatsHandler) GetStats(c *gin.Context) {
	result := make(map[string]interface{})

	// статистика пулл реквестов
	prStats, err := h.prService.GetStats()
	if err != nil {
		h.logger.Warnf("failed to get PR stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get PR stats"})
		return
	}
	result["prs"] = prStats

	// пользватели
	userStats, err := h.userService.GetStats()
	if err != nil {
		h.logger.Warnf("failed to get user stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user stats"})
		return
	}
	result["users"] = userStats

	// команды
	teamStats, err := h.teamService.GetStats()
	if err != nil {
		h.logger.Warnf("failed to get team stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get team stats"})
		return
	}
	result["teams"] = teamStats

	c.JSON(http.StatusOK, result)
}
