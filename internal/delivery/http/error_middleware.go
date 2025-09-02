package http

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	apperrors "github.com/russo2642/renti_kz/pkg/errors"
)


func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()


		if len(c.Errors) > 0 {

			err := c.Errors.Last().Err


			appErr, ok := apperrors.AsAppError(err)
			if !ok {

				appErr = apperrors.NewInternalError("Внутренняя ошибка сервера", err)
			}


			log.Printf("Error: %s", apperrors.FormatError(appErr))


			c.JSON(appErr.StatusCode, domain.NewErrorResponse(appErr.Message))
			c.Abort()
		}
	}
}


func HandleError(c *gin.Context, err error) {
	c.Error(err)
}


func RespondWithError(c *gin.Context, err error) {

	appErr, ok := apperrors.AsAppError(err)
	if !ok {

		appErr = apperrors.NewInternalError("Внутренняя ошибка сервера", err)
	}


	log.Printf("Error: %s", apperrors.FormatError(appErr))


	c.JSON(appErr.StatusCode, domain.NewErrorResponse(appErr.Message))
}
