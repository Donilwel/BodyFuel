package v1

import (
	"backend/internal/handlers/v1/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (a *API) registerAvatarsHandlers(router *gin.RouterGroup) {
	group := router.Group("/avatars")
	group.POST("", a.presignAvatar)
}

// presignAvatar Скидывает ссылку для загрузки аватарки пользователя
// @Summary Обработка аватарки
// @Description Генерирует и скидывает ссылку для загрузки аватарки пользователя со стороны клиента
// @Tags Photo
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.PresignAvatarRequest true "Request body"
// @Success 200 {object} models.PresignAvatarResponse "Успешная выдача публичной ссылки для загрузки"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} models.ErrorResponse "Неверные учетные данные"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /avatars [post]
func (a *API) presignAvatar(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("avatars error: presign avatar: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"avatars error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("avatars error: presign avatar: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"avatars error": "invalid user_id type in context",
		})
		return
	}

	var req models.PresignAvatarRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"avatars error": err.Error()})
		return
	}

	if err := a.validator.Struct(req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"avatars error": err.Error()})
		return
	}

	uploadURL, objectKey, err := a.avatarService.PresignPutAvatar(
		ctx.Request.Context(),
		userIDStr,
		req.ContentType,
	)
	if err != nil {
		a.log.Errorf("avatars error: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"avatars error": "cannot presign avatar"})
		return
	}

	avatarURL := a.avatarService.PublicAvatarURL(objectKey)

	a.log.Infof("avatars: presigned avatar link showed success: %s", avatarURL)
	ctx.JSON(http.StatusOK, models.PresignAvatarResponse{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		AvatarURL: avatarURL,
	})
}
