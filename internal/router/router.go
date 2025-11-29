package router

import (
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/handler"
	"github.com/gin-gonic/gin"
)

// NewRouter создаёт gin с роутами
func NewRouter(
	userH *handler.UserHandler,
	teamH *handler.TeamHandler,
	prH *handler.PullRequestHandler,
	statsH *handler.StatsHandler,
) *gin.Engine {
	router := gin.Default()

	// health-check
	router.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// пользователи
	router.POST("/users/setIsActive", userH.SetIsActive)
	router.GET("/users/getReview", userH.GetReviewPR)

	// команды
	router.POST("/team/add", teamH.CreateTeam)
	router.GET("/team/get", teamH.GetTeam)
	router.POST("/team/deactivate", teamH.DeactivateTeam)

	// пулл реквесты
	router.POST("/pullRequest/create", prH.CreatePR)
	router.POST("/pullRequest/merge", prH.MergePR)
	router.POST("/pullRequest/reassign", prH.ReassignReviewer)

	// статистика
	router.GET("/stats", statsH.GetStats)

	return router
}
