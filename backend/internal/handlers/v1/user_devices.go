package v1

import (
	"backend/internal/domain/entities"
	"backend/internal/handlers/v1/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (a *API) registerUserDevicesHandlers(router *gin.RouterGroup) {
	devices := router.Group("/user/devices")
	devices.POST("", a.registerDevice)
	devices.GET("", a.listDevices)
	devices.DELETE("/:uuid", a.deleteDevice)
}

// registerDevice регистрирует device token пользователя для пуш-уведомлений
// @Summary Регистрация device token
// @Description Регистрирует или обновляет device token устройства для получения APNs push-уведомлений
// @Tags Devices
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.RegisterDeviceRequest true "Device token и платформа"
// @Success 200 {object} models.UserDeviceResponse "Зарегистрированное устройство"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/devices [post]
func (a *API) registerDevice(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	var req models.RegisterDeviceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		a.log.Errorf("register device: invalid request: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := a.validator.Struct(req); err != nil {
		a.handleValidationErrors(ctx, err, "register device")
		return
	}

	if err := a.CRUDService.RegisterUserDevice(ctx, entities.UserDeviceInitSpec{
		UserID:      userID,
		DeviceToken: req.DeviceToken,
		Platform:    req.Platform,
	}); err != nil {
		a.log.Errorf("register device: internal error: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to register device"})
		return
	}

	a.log.Infof("register device: success for user %s", userID)
	ctx.JSON(http.StatusOK, gin.H{"message": "Device registered successfully"})
}

// listDevices возвращает список зарегистрированных устройств пользователя
// @Summary Список устройств пользователя
// @Description Возвращает все зарегистрированные устройства для push-уведомлений
// @Tags Devices
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.UserDeviceResponse "Список устройств"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/devices [get]
func (a *API) listDevices(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	devices, err := a.CRUDService.ListUserDevices(ctx, userID)
	if err != nil {
		a.log.Errorf("list devices: internal error: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to list devices"})
		return
	}

	ctx.JSON(http.StatusOK, models.NewUserDevicesResponse(devices))
}

// deleteDevice удаляет зарегистрированное устройство
// @Summary Удаление device token
// @Description Удаляет зарегистрированный device token по ID
// @Tags Devices
// @Security BearerAuth
// @Produce json
// @Param uuid path string true "ID устройства"
// @Success 200 {object} models.SuccessResponse "Успешно удалено"
// @Failure 400 {object} models.ErrorResponse "Неверный формат UUID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/devices/{uuid} [delete]
func (a *API) deleteDevice(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	deviceID, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid device id format"})
		return
	}

	if err := a.CRUDService.DeleteUserDevice(ctx, deviceID, userID); err != nil {
		a.log.Errorf("delete device: internal error: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete device"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Device deleted successfully"})
}
