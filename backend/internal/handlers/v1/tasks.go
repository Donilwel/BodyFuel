package v1

import (
	"backend/internal/dto"
	"backend/internal/handlers/v1/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (a *API) registerTasksHandlers(router *gin.RouterGroup) {
	task := router.Group("/tasks")
	task.GET("", a.listTasks)
	task.GET("/:uuid", a.getTask)
	task.DELETE("/:uuid", a.deleteTask)
	task.POST("/:uuid/restart", a.restartTask)
}

// getTask возвращает задачу по ID
// @Summary Получение задачи по ID
// @Description Возвращает одну задачу по её UUID
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Param uuid path string true "ID задачи"
// @Success 200 {object} models.TaskResponse "Задача"
// @Failure 400 {object} models.ErrorResponse "Неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 404 {object} models.ErrorResponse "Задача не найдена"
// @Router /tasks/{uuid} [get]
func (a *API) getTask(ctx *gin.Context) {
	taskID, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	task, err := a.CRUDService.GetTask(ctx, taskID)
	if err != nil {
		a.log.Errorf("get task error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	ctx.JSON(http.StatusOK, models.NewTaskResponse(task))
}

// listTasks возвращает список задач в очереди
// @Summary Список задач
// @Description Возвращает список задач в очереди executor'а
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.TaskResponse "Список задач"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /tasks [get]
func (a *API) listTasks(ctx *gin.Context) {
	tasks, err := a.CRUDService.ListTasks(ctx, dto.TasksFilter{})
	if err != nil {
		a.log.Errorf("list tasks error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to list tasks"})
		return
	}

	ctx.JSON(http.StatusOK, models.NewTasksResponse(tasks))
}

// restartTask перезапускает упавшую задачу
// @Summary Перезапуск задачи
// @Description Сбрасывает счётчик попыток и переводит задачу обратно в running
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Param uuid path string true "ID задачи"
// @Success 200 {object} models.SuccessResponse "Задача перезапущена"
// @Failure 400 {object} models.ErrorResponse "Неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /tasks/{uuid}/restart [post]
func (a *API) restartTask(ctx *gin.Context) {
	id := ctx.Param("uuid")
	if id == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing task id in path"})
		return
	}

	taskID, err := uuid.Parse(id)
	if err != nil {
		a.log.Errorf("restart task error: invalid uuid: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id format", "details": err.Error()})
		return
	}

	if err = a.CRUDService.RestartTask(ctx, taskID); err != nil {
		a.log.Errorf("restart task error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to restart task"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task restarted"})
}

// deleteTask удаляет задачу из очереди
// @Summary Удаление задачи
// @Description Удаляет задачу из очереди executor'а
// @Tags Tasks
// @Security BearerAuth
// @Produce json
// @Param uuid path string true "ID задачи"
// @Success 200 {object} models.SuccessResponse "Задача удалена"
// @Failure 400 {object} models.ErrorResponse "Неверный формат ID"
// @Failure 401 {object} models.ErrorResponse "Отсутствует авторизация"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /tasks/{uuid} [delete]
func (a *API) deleteTask(ctx *gin.Context) {
	id := ctx.Param("uuid")
	if id == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing task id in path"})
		return
	}

	taskID, err := uuid.Parse(id)
	if err != nil {
		a.log.Errorf("delete task error: invalid uuid: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id format", "details": err.Error()})
		return
	}

	if err = a.CRUDService.DeleteTask(ctx, taskID); err != nil {
		a.log.Errorf("delete task error: %s", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete task"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}
