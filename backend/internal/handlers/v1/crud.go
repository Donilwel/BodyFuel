package v1

import (
	"backend/internal/dto"
	"backend/internal/handlers/v1/models"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"net/http"
)

func (a *API) registerCRUDHandlers(router *gin.RouterGroup) {
	group := router.Group("/crud")

	user := group.Group("/user")
	user.DELETE("/info", a.deleteUserInfo)
	user.PATCH("/info", a.updateUserInfo)
	user.GET("/info", a.getUserInfo)

	user.GET("/params", a.getUserParams)
	user.PATCH("/params", a.updateUserParams)
	user.DELETE("/params", a.deleteUserParams)
	user.POST("/params", a.createUserParams)
}

func (a *API) deleteUserInfo(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: delete user info: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: delete user info: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: delete user info: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	err = a.CRUDService.DeleteInfoUser(ctx, dto.UserInfoFilter{ID: &userID})
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: delete user info: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully deleted"})
}

func (a *API) updateUserInfo(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: update user info: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: update user info: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: update user info: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	var m models.UserInfoUpdateRequestModel
	if err := ctx.ShouldBindJSON(&m); err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	if err := a.validator.Struct(m); err != nil {
		a.handleValidationAuthFields(ctx, err, "update user info")
		return
	}

	err = a.CRUDService.UpdateInfoUser(ctx, dto.UserInfoFilter{ID: &userID}, m.ToParam())
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: update user info: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully updated"})
}

func (a *API) getUserInfo(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: get user info: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: get user info: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: get user info: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	ui, err := a.CRUDService.GetInfoUser(ctx, dto.UserInfoFilter{ID: &userID}, false)
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: get user info: success")
	ctx.JSON(http.StatusOK, models.NewUserInfoResponse(ui))
}

func (a *API) getUserParams(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: get user params: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: get user params: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: get user params: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	up, err := a.CRUDService.GetParamsUser(ctx, dto.UserParamsFilter{UserID: &userID}, false)
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: get user params: success")
	ctx.JSON(http.StatusOK, models.NewUserParamsResponse(up))
}

func (a *API) updateUserParams(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: update user params: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: update user params: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: update user params: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	var m models.UserParamsUpdateRequestModel
	if err := ctx.ShouldBindJSON(&m); err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	if err := a.validator.Struct(m); err != nil {
		a.handleValidationErrors(ctx, err, "update user params")
		return
	}

	spec, err := m.ToParam()
	if err != nil {
		a.log.Errorf("crud error: invalid data: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: invalid data": err.Error()})
		return
	}

	err = a.CRUDService.UpdateParamsUser(ctx, dto.UserParamsFilter{UserID: &userID}, spec)
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: update user params: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully updated"})
}

func (a *API) handleValidationErrors(c *gin.Context, err error, contextKey string) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		a.log.Errorf("crud error: %s: %s", contextKey, err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error": "Internal validation error"})
		return
	}

	out := make(map[string]string)
	for _, fe := range ve {
		field := fe.Field()
		tag := fe.Tag()

		switch tag {
		case "required":
			out[field] = field + " is required"
		case "min":
			out[field] = field + " is too small"
		case "max":
			out[field] = field + " is too big"
		case "oneof":
			out[field] = field + " is invalid"
		default:
			out[field] = field + " is invalid"
		}
	}

	response := gin.H{
		"crud error": gin.H{
			contextKey: out,
		},
	}
	a.log.Errorf("crud error: validation error: %s", response)
	c.AbortWithStatusJSON(http.StatusBadRequest, response)
}

func (a *API) deleteUserParams(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: delete user params: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: delete user params: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: delete user params: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	err = a.CRUDService.DeleteParamsUser(ctx, dto.UserParamsFilter{UserID: &userID})
	if err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: delete user params: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully deleted"})
}

func (a *API) createUserParams(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("crud error: create user params: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("crud error: create user params: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("crud error: create user params: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	var m models.UserParamsCreateRequestModel
	if err := ctx.ShouldBindJSON(&m); err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	if err := a.validator.Struct(m); err != nil {
		a.handleValidationErrors(ctx, err, "create user params")
		return
	}
	up, err := m.ToSpec()
	if err != nil {
		a.log.Errorf("crud error: create user params: invalid data: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: create user params: invalid data": err.Error()})
		return
	}
	up.UserID = userID

	if err := a.CRUDService.CreateParamsUser(ctx, up); err != nil {
		a.log.Errorf("crud error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"crud error: internal error": err.Error()})
		return
	}

	a.log.Infof("crud info: create user params: success")
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully created"})
}
