package v1

import (
	"github.com/gin-gonic/gin"
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

	user.GET("/weight", a.getUserWeight)
	user.GET("/weight/history", a.getUserWeightHistory)
	user.PATCH("/weight", a.updateUserWeight)
	user.DELETE("/weight/:uuid", a.deleteUserWeight)
	user.POST("/weight", a.createUserWeight)

	task := group.Group("/tasks")
	task.DELETE("/:uuid", a.deleteTask)
	task.POST("/:uuid/restart", a.restartTask)
	task.GET("", a.listTasks)

	workout := group.Group("/workouts")
	workout.GET("/:uuid", a.getUserWorkout)
	workout.DELETE("/:uuid", a.deleteUserWorkout)
	workout.PATCH("/:uuid", a.updateUserWorkout)
	workout.POST("", a.createUserWorkout)
	workout.GET("/history", a.getUserWorkouts)

	exercises := workout.Group("/exercises")
	exercises.GET("/:uuid", a.getExercise)
	exercises.DELETE("/:uuid", a.deleteExercise)
	exercises.PATCH("/:uuid", a.updateExercise)
	exercises.POST("/:uuid", a.createExercise)
	exercises.GET("/history", a.getExercises)
}
