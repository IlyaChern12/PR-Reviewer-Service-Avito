package handler

import (
	"net/http"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	CodeTeamExists   = "TEAM_EXISTS"
	CodeTeamNotFound = "NOT_FOUND"
)

type TeamHandler struct {
	teamService *service.TeamService
	prService   *service.PullRequestService
	userService *service.UserService
	logger      *zap.SugaredLogger
}

func NewTeamHandler(teamService *service.TeamService, prService *service.PullRequestService, userService *service.UserService, logger *zap.SugaredLogger) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
		userService: userService,
		prService:   prService,
		logger:      logger,
	}
}

/*
	  создание команды
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
			400 TEAM_EXISTS: { "error": { "code": "TEAM_EXISTS", "message": "team already exists" } }
*/
func (h *TeamHandler) CreateTeam(ctx *gin.Context) {
	var req domain.Team

	// читаем тело
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    CodeInvalidInput,
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
					"code":    CodeTeamExists,
					"message": "team_name already exists",
				},
			})
			return
		}

		// 500 остальные ошибки
		h.logger.Warnf("failed to create team %s: %v", req.TeamName, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    CodeUnknownError,
				"message": err.Error(),
			},
		})
		return
	}

	team, err := h.teamService.GetTeam(req.TeamName)
	if err != nil {
		h.logger.Warnf("failed to fetch created team %s: %v", req.TeamName, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    CodeUnknownError,
				"message": err.Error(),
			},
		})
		return
	}

	// 201 успех
	h.logger.Infof("team %s created", req.TeamName)
	ctx.JSON(http.StatusCreated, gin.H{
		"team": team,
	})
}

/*
	 получение команды
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
		404 NOT_FOUND: { "error": { "code": "NOT_FOUND", "message": "team not found" } }
*/
func (h *TeamHandler) GetTeam(ctx *gin.Context) {
	teamName := ctx.Query("team_name")

	// проверка что введено имя
	if teamName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    CodeInvalidInput,
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
					"code":    CodeTeamNotFound,
					"message": "resource not found",
				},
			})
			return
		}

		// 500 другие ошибки
		h.logger.Warnf("failed to get team %s: %v", teamName, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    CodeUnknownError,
				"message": err.Error(),
			},
		})
		return
	}

	// 200 успех
	ctx.JSON(http.StatusOK, team)
}

/*
	 деактивация команды и переназначение ревьюверов
		POST /team/deactivate
		Body:
			{
				"team_name": "team1"
			}
		Response:
			200 OK:
				{ "message": "team deactivated and PR reviewers reassigned" }
			400 BAD_REQUEST:
				{ "error": "team_name is empty" } запроса пустое или некорректное
			404 NOT_FOUND:
				{ "error": "team not found" } указанной команды не существует
			500 INTERNAL_SERVER_ERROR:
				{ "error": err.Error() } внутренняя ошибка сервиса
*/
func (h *TeamHandler) DeactivateTeam(c *gin.Context) {
    // структура для чтения JSON body запроса
    var req struct {
        TeamName string `json:"team_name"` // имя команды, которую нужно деактивировать
    }

    // парсим JSON тело запроса и проверяем, что team_name не пустой
    if err := c.ShouldBindJSON(&req); err != nil || req.TeamName == "" {
        // если JSON некорректный или поле пустое — возвращаем 400
        c.JSON(http.StatusBadRequest, gin.H{"error": "team_name is empty"})
        return
    }

    teamName := req.TeamName // сохраняем для удобства

    // проверяем, существует ли команда
    exists, err := h.teamService.TeamExists(teamName)
    if err != nil {
        // если произошла ошибка при проверке — возвращаем 500
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if !exists {
        // если команды нет — возвращаем 404
        c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
        return
    }

    // массовая деактивация всех пользователей команды
    if err := h.userService.DeactivateTeam(teamName); err != nil {
        // если произошла ошибка при деактивации — возвращаем 500
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // безопасное переназначение ревьюверов для всех открытых PR команды
    if err := h.prService.ReassignReviewersForTeam(teamName); err != nil {
        // если ошибка при переназначении — возвращаем 500
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // успешный ответ: команда деактивирована и ревьюверы переназначены
    c.JSON(http.StatusOK, gin.H{"message": "team deactivated and PR reviewers reassigned"})
}