package handler

import (
	"net/http"

	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/domain"
	"github.com/IlyaChern12/PR-Reviewer-Service-Avito/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	CodeNotFound      = "NOT_FOUND"
	CodePRExists      = "PR_EXISTS"
	CodePRMerged      = "PR_MERGED"
	CodeNotAssigned   = "NOT_ASSIGNED"
	CodeNoCandidate   = "NO_CANDIDATE"
)

type PullRequestHandler struct {
	prService *service.PullRequestService
	logger    *zap.SugaredLogger
}

func NewPullRequestHandler(prService *service.PullRequestService, logger *zap.SugaredLogger) *PullRequestHandler {
	return &PullRequestHandler{
		prService: prService,
		logger:    logger,
	}
}



/*  создание пулл реквеста
	POST /pullRequest/create
	Body:
		{
			"pull_request_id": "pr-1",
			"pull_request_name": "Add new",
			"author_id": "user1"
		}
	Success
		201: { "pr": PullRequest }
	Errors:
		404 NOT_FOUND - автор или команда не найдены
		409 PR_EXISTS - PR с этим id уже есть
		400 INVALID_INPUT - некорректное тело запроса */
func (h *PullRequestHandler) CreatePR(c *gin.Context) {
	var req struct {
		ID   string `json:"pull_request_id"`
		Name string `json:"pull_request_name"`
		AuthorID string `json:"author_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{ "code": CodeInvalidInput, "message": err.Error() },
		})
		return
	}

	// сброка пулл реквеста
	pr := &domain.PullRequest{
		PRID:   req.ID,
		PRName: req.Name,
		AuthorID: req.AuthorID,
	}

	err := h.prService.CreatePR(pr)
	if err != nil {
		switch err {
		case service.ErrPRExists:
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{ "code": CodePRExists, "message": "PR already exists" },
			})
			return

		case service.ErrAuthorNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{ "code": CodeNotFound, "message": "author not found" },
			})
			return
		}

		// unknown
		h.logger.Warnf("failed to create PR %s: %v", req.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{ "code": CodeUnknownError, "message": err.Error() },
		})
		return
	}

	// успешное создание
	c.JSON(http.StatusCreated, gin.H{"pr": pr})
}



/*  слияние
	POST /pullRequest/merge
	Body:
		{ "pull_request_id": "pr-1" }
	Success 200:
		{ "pr": PullRequest (MERGED) }
	Errors:
		404 NOT_FOUND - PR не найден
		400 INVALID_INPUT - некорректное тело запроса */
func (h *PullRequestHandler) MergePR(c *gin.Context) {
	var req struct {
		ID string `json:"pull_request_id"`
	}

	// читаем JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    CodeInvalidInput,
				"message": err.Error(),
			},
		})
		return
	}

	// выполняем merge через сервис
	pr, err := h.prService.MergePR(req.ID)
	if err != nil {

		// PR не существует
		if err == service.ErrPRNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    CodeNotFound,
					"message": "PR not found",
				},
			})
			return
		}

		// unexpected
		h.logger.Warnf("failed to merge PR %s: %v", req.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    CodeUnknownError,
				"message": err.Error(),
			},
		})
		return
	}

	// успех
	c.JSON(http.StatusOK, gin.H{"pr": pr})
}



/* переназначение ревьюера
	POST /pullRequest/reassign
	Body:
		{
			"pull_request_id": "pr-1",
			"old_user_id": "user2"
		}
	Success 200:
		{
		"pr": { PullRequest },
		"replaced_by": "user5"
		}
	Errors:
		404 NOT_FOUND - PR или пользователь не найдены
		409 PR_MERGED - нельзя переприсвоить после MERGED
		409 NOT_ASSIGNED - переданный пользователь не ревьюер
		409 NO_CANDIDATE - нет активного пользователя для замены в команде */
func (h *PullRequestHandler) ReassignReviewer(c *gin.Context) {
	var req struct {
		PRID       string `json:"pull_request_id"`
		OldUserID  string `json:"old_user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": CodeInvalidInput, "message": err.Error()}})
		return
	}

	pr, newReviewerID, err := h.prService.ReassignReviewer(req.PRID, req.OldUserID)
	if err != nil {
		switch err {
		case service.ErrPRMerged:
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": CodePRMerged, "message": "cannot reassign on merged PR"}})
			return
		case service.ErrNoCandidate:
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": CodeNoCandidate, "message": "no active replacement candidate"}})
			return
		case service.ErrPRNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": CodeNotFound, "message": "PR not found"}})
			return
		default:
			h.logger.Warnf("failed to reassign reviewer on PR %s: %v", req.PRID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": CodeUnknownError, "message": err.Error()}})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"pr": pr,
		"replaced_by": newReviewerID,
	})
}