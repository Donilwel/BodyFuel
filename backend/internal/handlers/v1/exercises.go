package v1

import (
	"backend/internal/dto"
	"backend/internal/handlers/v1/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func (a *API) registerExerciseHandlers(router *gin.RouterGroup) {
	exercises := router.Group("/exercises")
	exercises.GET("/:uuid", a.getExercise)
	exercises.DELETE("/:uuid", a.deleteExercise)
	exercises.PATCH("/:uuid", a.updateExercise)
	exercises.POST("/:uuid", a.createExercise)
	exercises.GET("", a.getExercises)
}

func (a *API) getExercise(ctx *gin.Context) {
	//userIDRaw, ok := ctx.Get("user_id")
	//if !ok {
	//	a.log.Errorf("user weight error: get user weight: missing user_id in context")
	//	ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
	//		"error": "missing user_id in context",
	//	})
	//	return
	//}
	//
	//userIDStr, ok := userIDRaw.(string)
	//if !ok {
	//	a.log.Errorf("user weight error: get user weight: invalid user_id type in context")
	//	ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
	//		"error": "invalid user_id type in context",
	//	})
	//	return
	//}
	//
	//userID, err := uuid.Parse(userIDStr)
	//if err != nil {
	//	a.log.Errorf("user weight error: get user weight: invalid user_id format: %s", err.Error())
	//	ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
	//		"error":   "invalid user_id format",
	//		"details": err.Error(),
	//	})
	//	return
	//}
	//
	//uw, err := a.CRUDService.GetExercise(ctx, dto.ExerciseFilter{ID: id}, false)
	//if err != nil {
	//	a.log.Errorf("user weight error: internal error: %s", err.Error())
	//	ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"crud error: internal error": err.Error()})
	//	return
	//}
	//
	//a.log.Infof("user weight: get user weight: success")
	//ctx.JSON(http.StatusOK, models.NewUserWeightResponse(uw))
}

func (a *API) getExercises(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("exercise error: get exercises: missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		a.log.Errorf("exercise error: get exercises history: invalid user_id type in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id type in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		a.log.Errorf("exercise error: get exercises: invalid user_id format: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user_id format",
			"details": err.Error(),
		})
		return
	}

	le, err := a.CRUDService.ListExercise(ctx, userID, dto.ExerciseFilter{}, false)
	if err != nil {
		a.log.Errorf("exercise error: internal error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"exercise error: internal error": err.Error()})
		return
	}

	a.log.Infof("exercise: get exercises: success")
	ctx.JSON(http.StatusOK, models.NewExerciseResponseList(le))
}

func (a *API) updateExercise(ctx *gin.Context) {

}

func (a *API) deleteExercise(ctx *gin.Context) {

}

func (a *API) createExercise(ctx *gin.Context) {

}
