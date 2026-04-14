package v1

import (
	"backend/internal/domain/entities"
	"backend/internal/handlers/v1/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (a *API) registerNutritionHandlers(router *gin.RouterGroup) {
	group := router.Group("/nutrition")
	group.POST("/analyze/upload", a.uploadAndAnalyzeNutritionPhoto)
	group.POST("/entries", a.createFoodEntry)
	group.PATCH("/entries/:uuid", a.updateFoodEntry)
	group.DELETE("/entries/:uuid", a.deleteFoodEntry)
	group.GET("/diary", a.getNutritionDiary)
	group.GET("/report", a.getNutritionReport)
	group.GET("/recipes", a.recommendRecipes)
}

// uploadAndAnalyzeNutritionPhoto загружает фото еды, анализирует через OpenAI Vision и сохраняет запись в дневнике
// @Summary Загрузка фото еды с автоматическим добавлением в дневник
// @Description Принимает файл (JPEG/PNG/WebP) + тип приёма пищи. Загружает в S3, определяет КБЖУ через OpenAI Vision и автоматически создаёт запись в дневнике питания. Форматы HEIC/HEIF не поддерживаются.
// @Tags Nutrition
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param photo formData file true "Файл изображения еды (JPEG, PNG, WebP)"
// @Param meal_type formData string true "Тип приёма пищи" Enums(breakfast, lunch, dinner, snack)
// @Param date formData string false "Дата приёма пищи (YYYY-MM-DD), по умолчанию сегодня"
// @Success 201 {object} models.FoodEntryResponse "Созданная запись в дневнике"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /nutrition/analyze/upload [post]
func (a *API) uploadAndAnalyzeNutritionPhoto(ctx *gin.Context) {
	userID, err := a.getUserIDFromContext(ctx)
	if err != nil {
		return
	}

	file, header, err := ctx.Request.FormFile("photo")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "photo file is required"})
		return
	}
	defer file.Close()

	mealType := ctx.PostForm("meal_type")
	switch mealType {
	case "breakfast", "lunch", "dinner", "snack":
	case "":
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "meal_type is required (breakfast, lunch, dinner, snack)"})
		return
	default:
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "meal_type must be one of: breakfast, lunch, dinner, snack"})
		return
	}

	date := time.Now()
	if dateStr := ctx.PostForm("date"); dateStr != "" {
		parsed, parseErr := time.Parse("2006-01-02", dateStr)
		if parseErr != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
		date = parsed
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}
	if contentType == "image/heic" || contentType == "image/heif" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "формат HEIC не поддерживается, используйте JPEG, PNG или WebP"})
		return
	}

	objectName := fmt.Sprintf("%s_%s", uuid.New().String(), header.Filename)

	uploadResult, err := a.nutritionService.UploadAndAnalyzePhoto(ctx, userID.String(), objectName, contentType, file)
	if err != nil {
		a.log.Errorf("nutrition: upload and analyze photo: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to upload and analyze photo"})
		return
	}

	entryID := uuid.New()
	now := time.Now()
	spec := entities.UserFoodInitSpec{
		ID:          entryID,
		UserID:      userID,
		Description: uploadResult.Analysis.Description,
		Calories:    uploadResult.Analysis.Calories,
		Protein:     uploadResult.Analysis.Protein,
		Carbs:       uploadResult.Analysis.Carbs,
		Fat:         uploadResult.Analysis.Fat,
		MealType:    entities.MealType(mealType),
		PhotoURL:    uploadResult.PhotoURL,
		Date:        date,
	}

	if err := a.nutritionService.CreateFoodEntry(ctx, spec); err != nil {
		a.log.Errorf("nutrition: create entry from photo: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to save food entry"})
		return
	}

	ctx.JSON(http.StatusCreated, models.FoodEntryResponse{
		ID:          entryID,
		Description: spec.Description,
		Calories:    spec.Calories,
		Protein:     spec.Protein,
		Carbs:       spec.Carbs,
		Fat:         spec.Fat,
		MealType:    mealType,
		PhotoURL:    spec.PhotoURL,
		Date:        date,
		CreatedAt:   now,
	})
}

// createFoodEntry создаёт запись в дневнике питания
// @Summary Создание записи о еде
// @Tags Nutrition
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateFoodEntryRequest true "Данные о еде"
// @Success 201 {object} models.SuccessResponse
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

// recommendRecipes возвращает рекомендованные рецепты на основе дневника питания за день
// @Summary Рекомендации рецептов
// @Description На основе уже съеденного за день ИИ предлагает 3-5 рецептов для следующего приёма пищи
// @Tags Nutrition
// @Security BearerAuth
// @Produce json
// @Param date query string false "Дата (YYYY-MM-DD), по умолчанию сегодня"
// @Success 200 {array} models.RecipeResponse "Список рекомендованных рецептов"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /nutrition/recipes [get]
func (a *API) recommendRecipes(ctx *gin.Context) {
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

	recipes, err := a.nutritionService.RecommendRecipes(ctx, userID, date)
	if err != nil {
		a.log.Errorf("nutrition: recommend recipes: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate recipe recommendations"})
		return
	}

	ctx.JSON(http.StatusOK, models.NewRecipeResponseList(recipes))
}

