package handler

import (
	"net/http"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	CodeUserNotFound = "NOT_FOUND"
	CodeInvalidInput = "INVALID_INPUT"
	CodeUnknownError = "UNKNOWN_ERROR"
)

type UserHandler struct {
	userService *service.UserService
	logger      *zap.SugaredLogger
}

func NewUserHandler(userService *service.UserService, logger *zap.SugaredLogger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

/*
	  изменение статуса пользователя
		POST /users/setIsActive
		Body:
		{
			"user_id": "user",
			"is_active": false
		}
		Responses:
			200: { "user": {user object} }
			404: { "error": { "code": "NOT_FOUND", "message": "user not found" } }
			400: { "error": { "code": "INVALID_INPUT", "message": "..." } }
*/
func (h *UserHandler) SetIsActive(ctx *gin.Context) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	// парсим тело
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    CodeInvalidInput,
				"message": err.Error(),
			},
		})
		return
	}

	// обновляем активность через сервис
	user, err := h.userService.SetIsActive(req.UserID, req.IsActive)
	if err != nil {
		code := CodeUserNotFound
		h.logger.Warnf("failed to set isActive for user %s: %v", req.UserID, err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    code,
				"message": err.Error(),
			},
		})
		return
	}

	// обновление
	h.logger.Infof("user %s isActive updated to %v", req.UserID, req.IsActive)
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

/*
	 возвращает PR, где пользователь назначен ревьювером
		GET /users/getReview?user_id=user1
		Parameters:
			user_id - идентификатор пользователя
		Responses:
			200: { "user_id": "user", "pull_requests": [PullRequest1...] }
			404: { "error": { "code": "NOT_FOUND", "message": "user not found" } }
*/
func (h *UserHandler) GetReviewPR(ctx *gin.Context) {
	userID := ctx.Query("user_id")

	prs, err := h.userService.ListReviewPR(userID)
	if err != nil {
		h.logger.Warnf("failed to get review PR for user %s: %v", userID, err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    CodeUserNotFound,
				"message": err.Error(),
			},
		})
		return
	}

	if prs == nil {
		prs = []*domain.PullRequestShort{}
	}

	// успешный ответ
	ctx.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"pull_requests": prs,
	})
}
