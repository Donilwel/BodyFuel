package v1

import "github.com/gin-gonic/gin"

func (a *API) registerWorkoutsHandlers(router *gin.RouterGroup) {
	workout := router.Group("/workouts")
	workout.GET("/:uuid", a.getUserWorkout)
	workout.DELETE("/:uuid", a.deleteUserWorkout)
	workout.PATCH("/:uuid", a.updateUserWorkout)
	workout.POST("", a.generateWorkout)
	workout.GET("/history", a.getUserWorkouts)
}

func (a *API) getUserWorkout(ctx *gin.Context) {

}

func (a *API) getUserWorkouts(ctx *gin.Context) {

}

func (a *API) updateUserWorkout(ctx *gin.Context) {

}

func (a *API) deleteUserWorkout(ctx *gin.Context) {

}

func (a *API) generateWorkout(ctx *gin.Context) {

}
