package v1

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/JWT"
	"backend/pkg/logging"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"regexp"
)

type (
	AuthService interface {
		Register(ctx context.Context, ua entities.UserInfoInitSpec) error
		Login(ctx context.Context, ui entities.UserAuthInitSpec) (string, error)
	}

	UserStatisticsService interface {
		CreateUserParams(ctx context.Context, up entities.UserParamsInitSpec) error
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
	}

	AvatarService interface {
		PresignPutAvatar(ctx context.Context, userID string, contentType string) (uploadURL string, objectKey string, err error)
		PublicAvatarURL(objectKey string) string
	}
)

type Config struct {
	AuthService           AuthService
	UserStatisticsService UserStatisticsService
	CRUDService           CRUDService
	AvatarService         AvatarService
	Validator             validator.Validate
	Log                   logging.Entry
}

type API struct {
	authService           AuthService
	userStatisticsService UserStatisticsService
	CRUDService           CRUDService
	avatarService         AvatarService
	validator             validator.Validate
	log                   logging.Entry
}

func NewHandlers(c Config) *API {
	return &API{
		authService:           c.AuthService,
		userStatisticsService: c.UserStatisticsService,
		CRUDService:           c.CRUDService,
		avatarService:         c.AvatarService,
		validator:             c.Validator,
		log:                   c.Log,
	}
}

func (a *API) RegisterHandlers(r *gin.RouterGroup) {
	a.registerAuthHandlers(r)

	protected := r.Group("", JWT.JWTAuthMiddleware())
	a.registerCRUDHandlers(protected)
	a.registerAvatarsHandlers(protected)

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
