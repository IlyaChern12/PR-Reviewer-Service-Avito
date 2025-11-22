package handler

import (
	"net/http"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	CodeTeamExists  = "TEAM_EXISTS"
	CodeTeamNotFound = "NOT_FOUND"
)

type TeamHandler struct {
	teamService *service.TeamService
	logger      *zap.SugaredLogger
}

func NewTeamHandler(teamService *service.TeamService, logger *zap.SugaredLogger) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
		logger:      logger,
	}
}

/*  создание команды
	POST /team/add
	Body:
		{
			"team_name": "team1",
			"members": [
				{ "user_id": "user1", "username": "Ilya", "is_active": true },
				{ "user_id": "user2", "username": "AnotherIlya", "is_active": true }
			]
		}
	Response:
		201 { "team": { team object } }
		400 TEAM_EXISTS: { "error": { "code": "TEAM_EXISTS", "message": "team already exists" } } */
func (h *TeamHandler) CreateTeam(ctx *gin.Context) {
	var req domain.Team

	// читаем тело
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code": CodeInvalidInput,
				"message": err.Error(),
			},
		})
		return
	}

	// обращаемся в сервис
	err := h.teamService.CreateTeam(&req, req.Members)
	if err != nil {
		// 400 команда существует
		if err == service.ErrTeamExists {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code": CodeTeamExists,
					"message": "team already exists",
				},
			})
			return
		}

		// 500 остальные ошибки
		h.logger.Warnf("failed to create team %s: %v", req.TeamName, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code": CodeUnknownError,
				"message": err.Error(),
			},
		})
		return
	}

	// 201 успех
	h.logger.Infof("team %s created", req.TeamName)
	ctx.JSON(http.StatusCreated, gin.H{
		"team": req,
	})
}



/* получение команды
	GET /team/get?team_name=team1
	Parameters:
		team_name (string, required)
	Response:
	200 {
		"team_name": "team1",
		"members": [
			{ "user_id": "user1", "username": "Ilya", "is_active": true },
			{ "user_id": "user2", "username": "AnotherIlya", "is_active": true }
		]
	}
	404 NOT_FOUND: { "error": { "code": "NOT_FOUND", "message": "team not found" } } */
func (h *TeamHandler) GetTeam(ctx *gin.Context) {
	teamName := ctx.Query("team_name")

	// проверка что введено имя
	if teamName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code": CodeInvalidInput,
				"message": "team name is empty",
			},
		})
		return
	}

	// идем в сервис
	team, err := h.teamService.GetTeam(teamName)
	if err != nil {
		// 404 команда не найдена
		if err == service.ErrTeamNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code": CodeTeamNotFound,
					"message": "team not found",
				},
			})
			return
		}

		// 500 другие ошибки
		h.logger.Warnf("failed to get team %s: %v", teamName, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code": CodeUnknownError,
				"message": err.Error(),
			},
		})
		return
	}

	// 200 успех
	ctx.JSON(http.StatusOK, team)
}

