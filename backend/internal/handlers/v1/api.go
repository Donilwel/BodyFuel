package v1

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/service/auth"
	"backend/internal/service/nutricion"
	"backend/pkg/JWT"
	"backend/pkg/ai"
	"backend/pkg/logging"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type (
	AuthService interface {
		Register(ctx context.Context, ua entities.UserInfoInitSpec) error
		Login(ctx context.Context, ui entities.UserAuthInitSpec) (auth.TokenPair, error)
		Refresh(ctx context.Context, rawToken string) (auth.TokenPair, error)
		SendVerificationCode(ctx context.Context, userID uuid.UUID, codeType entities.VerificationCodeType) error
		VerifyCode(ctx context.Context, userID uuid.UUID, code string, codeType entities.VerificationCodeType) error
		SendRecoveryCode(ctx context.Context, email string) error
		ResetPassword(ctx context.Context, email, code, newPassword string) error
	}

	UserStatisticsService interface {
		CreateUserParams(ctx context.Context, up entities.UserParamsInitSpec) error
	}

	WorkoutService interface {
		GenerateCustomWorkout(ctx context.Context, params *dto.GenerateWorkoutParams) (*entities.Workout, error)
	}

	CRUDService interface {
		GetInfoUser(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error)
		CreateInfoUser(ctx context.Context, info entities.UserInfoInitSpec) error
		UpdateInfoUser(ctx context.Context, f dto.UserInfoFilter, info entities.UserInfoUpdateParams) error
		DeleteInfoUser(ctx context.Context, f dto.UserInfoFilter) error

		GetParamsUser(ctx context.Context, f dto.UserParamsFilter, withBlock bool) (*entities.UserParams, error)
		CreateParamsUser(ctx context.Context, params entities.UserParamsInitSpec) error
		UpdateParamsUser(ctx context.Context, f dto.UserParamsFilter, userParams entities.UserParamsUpdateParams) error
		DeleteParamsUser(ctx context.Context, f dto.UserParamsFilter) error

		GetWeightUser(ctx context.Context, f dto.UserWeightFilter, withBlock bool) (*entities.UserWeight, error)
		ListWeightsUser(ctx context.Context, f dto.UserWeightFilter, withBlock bool) ([]*entities.UserWeight, error)
		CreateWeightUser(ctx context.Context, weight entities.UserWeightInitSpec) error
		UpdateWeightUser(ctx context.Context, f dto.UserWeightFilter, weight entities.UserWeightUpdateParams) error
		DeleteWeightUser(ctx context.Context, f dto.UserWeightFilter) error

		GetExercise(ctx context.Context, f dto.ExerciseFilter, withBlock bool) (*entities.Exercise, error)
		CreateExercise(ctx context.Context, params entities.ExerciseInitSpec) error
		UpdateExercise(ctx context.Context, f dto.ExerciseFilter, exercise entities.ExerciseUpdateParams) error
		DeleteExercise(ctx context.Context, f dto.ExerciseFilter) error
		ListExercise(ctx context.Context, userID uuid.UUID, f dto.ExerciseFilter, withBlock bool) ([]*entities.Exercise, error)

		GetWorkoutExercise(ctx context.Context, f dto.WorkoutsExerciseFilter, withBlock bool) (*entities.WorkoutsExercise, error)
		ListWorkoutsExercise(ctx context.Context, f dto.WorkoutsExerciseFilter) ([]*entities.WorkoutsExercise, error)
		CreateWorkoutExercise(ctx context.Context, workoutExercise *entities.WorkoutsExercise) error
		UpdateWorkoutExerciseByFilter(ctx context.Context, f dto.WorkoutsExerciseFilter, params entities.WorkoutsExerciseUpdateParams) error
		DeleteWorkoutExercise(ctx context.Context, f dto.WorkoutsExerciseFilter) error
		GetWorkout(ctx context.Context, f dto.WorkoutsFilter, withBlock bool) (*entities.Workout, error)
		ListWorkouts(ctx context.Context, f dto.WorkoutsFilter, withBlock bool) ([]*entities.Workout, error)
		UpdateWorkoutByFilter(ctx context.Context, f dto.WorkoutsFilter, params entities.WorkoutUpdateParams) error
		DeleteWorkout(ctx context.Context, f dto.WorkoutsFilter) error

		GetTask(ctx context.Context, id uuid.UUID) (*entities.Task, error)
		DeleteTask(ctx context.Context, id uuid.UUID) error
		ListTasks(ctx context.Context, filter dto.TasksFilter) ([]*entities.Task, error)
		RestartTask(ctx context.Context, id uuid.UUID) error

		RegisterUserDevice(ctx context.Context, spec entities.UserDeviceInitSpec) error
		ListUserDevices(ctx context.Context, userID uuid.UUID) ([]*entities.UserDevice, error)
		DeleteUserDevice(ctx context.Context, id, userID uuid.UUID) error

		CreateUserCalories(ctx context.Context, spec entities.UserCaloriesInitSpec) error
		GetUserCalories(ctx context.Context, f dto.UserCaloriesFilter) (*entities.UserCalories, error)
		ListUserCalories(ctx context.Context, f dto.UserCaloriesFilter) ([]*entities.UserCalories, error)
		UpdateUserCalories(ctx context.Context, f dto.UserCaloriesFilter, params entities.UserCaloriesUpdateParams) error
		DeleteUserCalories(ctx context.Context, id, userID uuid.UUID) error
	}

	AvatarService interface {
		PresignPutAvatar(ctx context.Context, userID string, contentType string) (uploadURL string, objectKey string, err error)
		PublicAvatarURL(objectKey string) string
	}

	NutritionService interface {
		AnalyzePhoto(ctx context.Context, imageURL string) (*ai.NutritionAnalysis, error)
		UploadAndAnalyzePhoto(ctx context.Context, userID, filename, contentType string, data io.Reader) (*nutricion.UploadPhotoResult, error)
		CreateFoodEntry(ctx context.Context, spec entities.UserFoodInitSpec) error
		GetFoodEntry(ctx context.Context, id, userID uuid.UUID) (*entities.UserFood, error)
		UpdateFoodEntry(ctx context.Context, id, userID uuid.UUID, params entities.UserFoodUpdateParams) error
		DeleteFoodEntry(ctx context.Context, id, userID uuid.UUID) error
		GetDiary(ctx context.Context, userID uuid.UUID, date time.Time) (*nutricion.NutritionDiary, error)
		GetReport(ctx context.Context, userID uuid.UUID, from, to time.Time) (*nutricion.NutritionReport, error)
		RecommendRecipes(ctx context.Context, userID uuid.UUID, date time.Time) ([]ai.RecipeItem, error)
	}

	RecommendationService interface {
		List(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entities.UserRecommendation, error)
		Refresh(ctx context.Context, userID uuid.UUID) ([]*entities.UserRecommendation, error)
		MarkRead(ctx context.Context, id, userID uuid.UUID) error
	}
)

type Config struct {
	AuthService           AuthService
	UserStatisticsService UserStatisticsService
	WorkoutService        WorkoutService
	CRUDService           CRUDService
	AvatarService         AvatarService
	NutritionService      NutritionService
	RecommendationService RecommendationService
	Validator             validator.Validate
	Log                   logging.Entry
}

type API struct {
	authService           AuthService
	userStatisticsService UserStatisticsService
	WorkoutService        WorkoutService
	CRUDService           CRUDService
	avatarService         AvatarService
	nutritionService      NutritionService
	recommendationService RecommendationService
	validator             validator.Validate
	log                   logging.Entry
}

func NewHandlers(c Config) *API {
	return &API{
		authService:           c.AuthService,
		userStatisticsService: c.UserStatisticsService,
		WorkoutService:        c.WorkoutService,
		CRUDService:           c.CRUDService,
		avatarService:         c.AvatarService,
		nutritionService:      c.NutritionService,
		recommendationService: c.RecommendationService,
		validator:             c.Validator,
		log:                   c.Log,
	}
}

func (a *API) RegisterHandlers(r *gin.RouterGroup) {
	a.registerAuthHandlers(r)

	protected := r.Group("", JWT.JWTAuthMiddleware())
	a.registerExerciseHandlers(protected)
	a.registerWorkoutsHandlers(protected)
	a.registerUserInfoHandlers(protected)
	a.registerUserParamsHandlers(protected)
	a.registerUserWeightHandlers(protected)
	a.registerTasksHandlers(protected)
	a.registerAvatarsHandlers(protected)
	a.registerUserDevicesHandlers(protected)
	a.registerUserCaloriesHandlers(protected)
	a.registerNutritionHandlers(protected)
	a.registerRecommendationsHandlers(protected)
}

func (a *API) checkPhone(ctx *gin.Context, phone string) error {
	phoneRegex := `^\+?[0-9]{10,15}$`
	re := regexp.MustCompile(phoneRegex)

	if !re.MatchString(phone) {
		return fmt.Errorf("uncorrect inpit phone number")
	}

	return nil
}

func (a *API) handleValidationAuthFields(c *gin.Context, err error, typeMethod string) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		a.log.Errorf("%s: %v", "auth error", "Internal error")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"auth error": "Internal error"})
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
			out[field] = field + " is too short"
		case "regex":
			out[field] = "phone number is invalid"
		case "email":
			out[field] = "email is invalid"
		default:
			out[field] = field + " is invalid"
		}
	}

	response := gin.H{
		"auth error": gin.H{
			typeMethod: out,
		},
	}
	a.log.Errorf("%s: %s: %v", "auth error", typeMethod, out)
	c.AbortWithStatusJSON(http.StatusBadRequest, response)
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

func (a *API) getUserIDFromContext(ctx *gin.Context) (uuid.UUID, error) {
	userIDRaw, ok := ctx.Get("user_id")
	if !ok {
		a.log.Errorf("missing user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing user_id in context",
		})
		return uuid.Nil, fmt.Errorf("missing user_id")
	}

	userID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		a.log.Errorf("invalid user_id in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user_id in context",
		})
		return uuid.Nil, fmt.Errorf("invalid user_id")
	}

	return userID, nil
}
