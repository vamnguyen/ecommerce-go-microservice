package router

import (
	"auth-service/internal/delivery/http/handler"
	"auth-service/internal/delivery/http/middleware"
	"auth-service/internal/infrastructure/config"
	"auth-service/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine         *gin.Engine
	authHandler    *handler.AuthHandler
	healthHandler  *handler.HealthHandler
	authMiddleware *middleware.AuthMiddleware
	logger         *logger.Logger
	config         *config.Config
}

func NewRouter(
	authHandler *handler.AuthHandler,
	healthHandler *handler.HealthHandler,
	authMiddleware *middleware.AuthMiddleware,
	logger *logger.Logger,
	config *config.Config,
) *Router {
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	return &Router{
		engine:         engine,
		authHandler:    authHandler,
		healthHandler:  healthHandler,
		authMiddleware: authMiddleware,
		logger:         logger,
		config:         config,
	}
}

func (r *Router) Setup() *gin.Engine {
	r.engine.Use(middleware.Recovery(r.logger, r.config.Environment))
	r.engine.Use(middleware.Logger(r.logger))
	r.engine.Use(middleware.CORS(r.config.Security.AllowedOrigins))

	r.setupRoutes()

	return r.engine
}

func (r *Router) setupRoutes() {
	v1 := r.engine.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.GET("/health", r.healthHandler.Health)
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/login", r.authHandler.Login)

			protected := auth.Group("")
			protected.Use(r.authMiddleware.RequireAuth())
			{
				protected.POST("/refresh", r.authHandler.RefreshToken)
				protected.POST("/logout", r.authHandler.Logout)
				protected.POST("/logout-all", r.authHandler.LogoutAll)
				protected.GET("/me", r.authHandler.GetMe)
				protected.PUT("/change-password", r.authHandler.ChangePassword)
			}
		}
	}
}
