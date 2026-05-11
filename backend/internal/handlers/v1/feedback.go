package v1

import (
	"backend/internal/handlers/v1/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *API) registerFeedbackHandlers(router *gin.RouterGroup) {
	feedback := router.Group("/feedback")
	feedback.POST("", a.sendFeedback)
}

// sendFeedback отправляет обратную связь на email
// @Summary Отправить обратную связь
// @Description Отправляет сообщение обратной связи на email администратора
// @Tags Feedback
// @Accept json
// @Produce json
// @Param request body models.SendFeedbackRequest true "Данные обратной связи"
// @Success 200 {object} models.FeedbackResponse "Сообщение отправлено"
// @Failure 400 {object} models.ErrorResponse "Неверный формат запроса"
// @Failure 500 {object} models.ErrorResponse "Ошибка отправки"
// @Router /feedback [post]
func (a *API) sendFeedback(ctx *gin.Context) {
	var req models.SendFeedbackRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request format", "details": err.Error()})
		return
	}

	subject := "BodyFuel Feedback"
	body := fmt.Sprintf("<h2>Новое сообщение обратной связи</h2>"+
		"<p><strong>Сообщение:</strong></p>"+
		"<p>%s</p>"+
		"<hr>"+
		"<p><em>Отправлено из приложения BodyFuel</em></p>", req.Message)

	if req.Email != "" {
		body += fmt.Sprintf("<p><strong>Email пользователя:</strong> %s</p>", req.Email)
	}

	adminEmail := "fantomrick228@gmail.com"

	if err := a.emailService.SendEmail(adminEmail, subject, body); err != nil {
		a.log.Errorf("send feedback error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to send feedback"})
		return
	}

	ctx.JSON(http.StatusOK, models.FeedbackResponse{
		Message: "Feedback sent successfully",
	})
}
