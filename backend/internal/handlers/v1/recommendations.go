package v1

import (
	"backend/internal/handlers/v1/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (a *API) registerRecommendationsHandlers(router *gin.RouterGroup) {
	group := router.Group("/recommendations")
	group.GET("", a.listRecommendations)
	group.POST("/refresh", a.refreshRecommendations)
	group.PATCH("/:uuid/read", a.markRecommendationRead)
}

// listRecommendations возвращает список рекомендаций пользователя
// @Summary Список рекомендаций
// @Tags Recommendations
// @Security BearerAuth
// @Produce json
// @Param page query int false "Страница (по умолчанию 1)"
// @Param limit query int false "Элементов на странице (по умолчанию 10)"
// @Success 200 {array} models.RecommendationResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /recommendations [get]
func (a *API) listRecommendations(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	recs, err := a.recommendationService.List(ctx, userID, page, limit)
	if err != nil {
		a.log.Errorf("recommendations: list: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to list recommendations"})
		return
	}

	ctx.JSON(http.StatusOK, models.NewRecommendationResponseList(recs))
}

// refreshRecommendations генерирует новые рекомендации через OpenAI
// @Summary Обновление рекомендаций
// @Description Генерирует персонализированные рекомендации через GPT, заменяя старые
// @Tags Recommendations
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.RecommendationResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /recommendations/refresh [post]
func (a *API) refreshRecommendations(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	recs, err := a.recommendationService.Refresh(ctx, userID)
	if err != nil {
		a.log.Errorf("recommendations: refresh: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh recommendations"})
		return
	}

	ctx.JSON(http.StatusOK, models.NewRecommendationResponseList(recs))
}

// markRecommendationRead отмечает рекомендацию как прочитанную
// @Summary Отметить рекомендацию прочитанной
// @Tags Recommendations
// @Security BearerAuth
// @Param uuid path string true "ID рекомендации"
// @Success 204
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /recommendations/{uuid}/read [patch]
func (a *API) markRecommendationRead(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	recID, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := a.recommendationService.MarkRead(ctx, recID, userID); err != nil {
		a.log.Errorf("recommendations: mark read: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to mark recommendation as read"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
