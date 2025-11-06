package middleware

import (
	"net/http"
	"runtime/debug"

	"auth-service/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(log *logger.Logger, environment string) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fields := []zap.Field{
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				}

				if environment == "development" {
					fields = append(fields, zap.ByteString("stack", debug.Stack()))
				}

				log.Error("Panic recovered", fields...)

				// Only write response if nothing has been written yet
				if !c.Writer.Written() {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "internal server error",
					})
				}
				c.Abort()
			}
		}()

		c.Next()
	}
}
