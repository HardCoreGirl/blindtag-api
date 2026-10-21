package router

import (
	"strings"
	"time"

	"blindtag_api/internal/config"
	"blindtag_api/internal/database"
	"blindtag_api/internal/handlers"
	"blindtag_api/internal/middleware"
	"blindtag_api/internal/repository"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes Gin engine with CORS, recovery, JWT auth middleware, and API routes.
func SetupRouter(cfg *config.Config, db *database.DB) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	if cfg.CORSAllowedOrigins == "*" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = strings.Split(cfg.CORSAllowedOrigins, ",")
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	corsConfig.MaxAge = 12 * time.Hour
	r.Use(cors.New(corsConfig))

	// Repositories
	userRepo := repository.NewUserRepository(db)
	scoreRepo := repository.NewScoreRepository(db)
	graffitiRepo := repository.NewGraffitiRepository(db)
	stageRepo := repository.NewStageRepository(db)

	// Handlers
	healthHandler := handlers.NewHealthHandler(db)
	userHandler := handlers.NewUserHandler(userRepo, cfg.JWTSecret)
	scoreHandler := handlers.NewScoreHandler(scoreRepo)
	graffitiHandler := handlers.NewGraffitiHandler(graffitiRepo)
	stageHandler := handlers.NewStageHandler(stageRepo)

	// Liveness & health check for GCP Cloud Run
	r.GET("/health", healthHandler.Check)
	r.GET("/", healthHandler.Check)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// 1. Users Group
		users := v1.Group("/users")
		{
			// Public: Guest login & account registration
			users.POST("/guest", userHandler.LoginGuest)
			// Public: View profile
			users.GET("/:id", userHandler.GetProfile)
			// Protected: Update own profile requires valid JWT token
			users.PUT("/:id", middleware.AuthMiddleware(cfg.JWTSecret), userHandler.UpdateProfile)
		}

		// 2. Scores & Leaderboards
		scores := v1.Group("/scores")
		{
			// Protected: Submitting game clear score requires authenticated runner token
			scores.POST("", middleware.AuthMiddleware(cfg.JWTSecret), scoreHandler.SubmitScore)
		}
		leaderboards := v1.Group("/leaderboard")
		{
			// Public: Everyone can view leaderboard rankings
			leaderboards.GET("/:stage_id", scoreHandler.GetLeaderboard)
		}

		// 3. Graffitis
		graffiti := v1.Group("/graffiti")
		{
			// Protected: Painting graffiti requires valid session token
			graffiti.POST("", middleware.AuthMiddleware(cfg.JWTSecret), graffitiHandler.CreateGraffiti)
			// Public: Viewing room wall decals
			graffiti.GET("", graffitiHandler.GetGraffitis)
			// Protected: Erasing graffiti requires authentication
			graffiti.DELETE("/:id", middleware.AuthMiddleware(cfg.JWTSecret), graffitiHandler.DeleteGraffiti)
		}

		// 4. Stages & Shared Altar Verification (Anti-Cheat & Shared Labyrinth Exits)
		stages := v1.Group("/stages")
		{
			// Protected: Check if current room matches the permanent shared exit in s_stage_config
			stages.POST("/check-altar", middleware.AuthMiddleware(cfg.JWTSecret), stageHandler.CheckAltar)
		}
	}

	return r
}
