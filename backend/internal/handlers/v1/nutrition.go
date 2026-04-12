package v1

import (
	"backend/internal/domain/entities"
	"backend/internal/handlers/v1/models"
	"backend/pkg/ai"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (a *API) registerNutritionHandlers(router *gin.RouterGroup) {
	group := router.Group("/nutrition")
	group.POST("/analyze", a.analyzeNutritionPhoto)
	group.POST("/entries", a.createFoodEntry)
	group.PATCH("/entries/:uuid", a.updateFoodEntry)
	group.DELETE("/entries/:uuid", a.deleteFoodEntry)
	group.GET("/diary", a.getNutritionDiary)
	group.GET("/report", a.getNutritionReport)
}

// analyzeNutritionPhoto анализирует фото еды через OpenAI Vision
// @Summary Анализ фото еды
// @Description Отправляет URL изображения еды в OpenAI Vision и возвращает питательную ценность
// @Tags Nutrition
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.AnalyzePhotoRequest true "URL изображения"
// @Success 200 {object} models.NutritionAnalysisResponse "Результат анализа"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /nutrition/analyze [post]
func (a *API) analyzeNutritionPhoto(ctx *gin.Context) {
	var m models.AnalyzePhotoRequest
	if err := ctx.ShouldBindJSON(&m); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := a.validator.Struct(m); err != nil {
		a.handleValidationErrors(ctx, err, "analyze photo")
		return
	}

	result, err := a.nutritionService.AnalyzePhoto(ctx, m.ImageURL)
	if err != nil {
		a.log.Errorf("nutrition: analyze photo: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to analyze photo"})
		return
	}

	ctx.JSON(http.StatusOK, models.NewNutritionAnalysisResponse(result))
}

// createFoodEntry создаёт запись в дневнике питания
// @Summary Создание записи о еде
// @Tags Nutrition
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateFoodEntryRequest true "Данные о еде"
// @Success 201 {object} models.FoodEntryResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /nutrition/entries [post]
func (a *API) createFoodEntry(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	var m models.CreateFoodEntryRequest
	if err := ctx.ShouldBindJSON(&m); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := a.validator.Struct(m); err != nil {
		a.handleValidationErrors(ctx, err, "create food entry")
		return
	}

	date := m.Date
	if date.IsZero() {
		date = time.Now()
	}

	spec := entities.UserFoodInitSpec{
		ID:          uuid.New(),
		UserID:      userID,
		Description: m.Description,
		Calories:    m.Calories,
		Protein:     m.Protein,
		Carbs:       m.Carbs,
		Fat:         m.Fat,
		MealType:    entities.MealType(m.MealType),
		PhotoURL:    m.PhotoURL,
		Date:        date,
	}

	if err := a.nutritionService.CreateFoodEntry(ctx, spec); err != nil {
		a.log.Errorf("nutrition: create entry: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create food entry"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Food entry created"})
}

// updateFoodEntry обновляет запись о еде
// @Summary Обновление записи о еде
// @Tags Nutrition
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param uuid path string true "ID записи"
// @Param request body models.UpdateFoodEntryRequest true "Поля для обновления"
// @Success 204
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /nutrition/entries/{uuid} [patch]
func (a *API) updateFoodEntry(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	entryID, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var m models.UpdateFoodEntryRequest
	if err := ctx.ShouldBindJSON(&m); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := a.validator.Struct(m); err != nil {
		a.handleValidationErrors(ctx, err, "update food entry")
		return
	}

	params := entities.UserFoodUpdateParams{
		Description: m.Description,
		Calories:    m.Calories,
		Protein:     m.Protein,
		Carbs:       m.Carbs,
		Fat:         m.Fat,
		PhotoURL:    m.PhotoURL,
		Date:        m.Date,
	}
	if m.MealType != nil {
		mt := entities.MealType(*m.MealType)
		params.MealType = &mt
	}

	if err := a.nutritionService.UpdateFoodEntry(ctx, entryID, userID, params); err != nil {
		a.log.Errorf("nutrition: update entry: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to update food entry"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// deleteFoodEntry удаляет запись о еде
// @Summary Удаление записи о еде
// @Tags Nutrition
// @Security BearerAuth
// @Param uuid path string true "ID записи"
// @Success 204
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /nutrition/entries/{uuid} [delete]
func (a *API) deleteFoodEntry(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	entryID, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := a.nutritionService.DeleteFoodEntry(ctx, entryID, userID); err != nil {
		a.log.Errorf("nutrition: delete entry: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete food entry"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// getNutritionDiary возвращает дневник питания за день
// @Summary Дневник питания за день
// @Tags Nutrition
// @Security BearerAuth
// @Produce json
// @Param date query string false "Дата (YYYY-MM-DD), по умолчанию сегодня"
// @Success 200 {object} models.NutritionDiaryResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /nutrition/diary [get]
func (a *API) getNutritionDiary(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	date := time.Now()
	if d := ctx.Query("date"); d != "" {
		parsed, err := time.Parse("2006-01-02", d)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
		date = parsed
	}

	diary, err := a.nutritionService.GetDiary(ctx, userID, date)
	if err != nil {
		a.log.Errorf("nutrition: get diary: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to get diary"})
		return
	}

	ctx.JSON(http.StatusOK, models.NewNutritionDiaryResponse(diary))
}

// getNutritionReport возвращает отчёт по питанию за период
// @Summary Отчёт по питанию за период
// @Tags Nutrition
// @Security BearerAuth
// @Produce json
// @Param from query string true "Начало периода (YYYY-MM-DD)"
// @Param to query string true "Конец периода (YYYY-MM-DD)"
// @Success 200 {object} models.NutritionReportResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /nutrition/report [get]
func (a *API) getNutritionReport(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	fromStr := ctx.Query("from")
	toStr := ctx.Query("to")
	if fromStr == "" || toStr == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "from and to query params are required"})
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid from format, use YYYY-MM-DD"})
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid to format, use YYYY-MM-DD"})
		return
	}

	report, err := a.nutritionService.GetReport(ctx, userID, from, to)
	if err != nil {
		a.log.Errorf("nutrition: get report: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to get report"})
		return
	}

	ctx.JSON(http.StatusOK, models.NewNutritionReportResponse(report))
}

// dummy type to satisfy interface reference in api.go
var _ = (*ai.NutritionAnalysis)(nil)
