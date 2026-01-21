package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

func (a *API) registerParamsHandlers(router *gin.RouterGroup) {
}

func (a *API) handleValidationUserParamsFields(c *gin.Context, err error, typeMethod string) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"params error": "Internal error"})
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
		case "e164":
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

	c.AbortWithStatusJSON(http.StatusBadRequest, response)
}
