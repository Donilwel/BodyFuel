package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (a *API) registerTasksHandlers(router *gin.RouterGroup) {
	task := router.Group("/tasks")
	task.DELETE("/:uuid", a.deleteTask)
	task.POST("/:uuid/restart", a.restartTask)
	task.GET("", a.listTasks)
}

func (a *API) listTasks(ctx *gin.Context) {
	//tasks, err := a.crudService.ListTasks(ctx.Request.Context(), dto.TasksFilter{})
	//if err != nil {
	//	ctx.Error(err)
	//	return
	//}

	//ctx.JSON(http.StatusOK, models.NewTasksResponse(tasks))
}

func (a *API) restartTask(ctx *gin.Context) {
	//id := ctx.Param("uuid")
	//if id == "" {
	//	ctx.Error(errs.ErrMissingUUIDParam())
	//	return
	//}

	//uuid, err := uuid.Parse(id)
	//if err != nil {
	//	ctx.Error(errs.ErrParsingUUID())
	//	return
	//}

	//if err = a.CRUDService.RestartTask(ctx.Request.Context(), uuid); err != nil {
	//	ctx.Error(err)
	//	return
	//}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task restarted"})
}

func (a *API) deleteTask(ctx *gin.Context) {
	//id := ctx.Param("uuid")
	//if id == "" {
	//	ctx.Error(errs.ErrMissingUUIDParam())
	//	return
	//}
	//
	//uuid, err := uuid.Parse(id)
	//if err != nil {
	//	ctx.Error(errs.ErrParsingUUID())
	//	return
	//}
	//
	//if err = a.CRUDService.DeleteTask(ctx.Request.Context(), uuid); err != nil {
	//	ctx.Error(err)
	//	return
	//}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}
