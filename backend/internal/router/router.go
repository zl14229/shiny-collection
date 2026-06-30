package router

import (
	"shiny-collection/internal/handler"
	"shiny-collection/internal/middleware"
	"shiny-collection/internal/repository"
	"shiny-collection/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, logger *zap.Logger, corsOrigins []string) *gin.Engine {
	r := gin.New()

	// global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(corsOrigins))
	r.Use(middleware.Logger(logger))

	// init repositories
	recordRepo := repository.NewRecordRepository(db)
	pokemonRepo := repository.NewPokemonRepository(db)
	gameRepo := repository.NewGameRepository(db)
	methodRepo := repository.NewMethodRepository(db)

	// init services
	recordSvc := service.NewRecordService(recordRepo)

	// init handlers
	recordH := handler.NewRecordHandler(recordSvc, logger)
	pokemonH := handler.NewPokemonHandler(pokemonRepo, logger)
	gameH := handler.NewGameHandler(gameRepo, logger)
	methodH := handler.NewMethodHandler(methodRepo, logger)
	statsH := handler.NewStatsHandler(recordSvc, logger)

	// API routes
	v1 := r.Group("/api/v1")
	{
		// records
		v1.GET("/records", recordH.List)
		v1.POST("/records", recordH.Create)
		v1.GET("/records/:id", recordH.Get)
		v1.PUT("/records/:id", recordH.Update)
		v1.DELETE("/records/:id", recordH.Delete)

		// video upload
		v1.POST("/records/:id/video", recordH.UploadVideo)
		v1.DELETE("/records/:id/video", recordH.DeleteVideo)

		// pokemon
		v1.GET("/pokemon", pokemonH.List)
		v1.GET("/pokemon/:id", pokemonH.Get)
		v1.POST("/pokemon", pokemonH.Create)

		// games
		v1.GET("/games", gameH.List)
		v1.GET("/games/:id", gameH.Get)

		// methods
		v1.GET("/methods", methodH.List)

		// stats
		v1.GET("/stats/overview", statsH.Overview)
		v1.GET("/stats/by-game", statsH.ByGame)
	}

	// serve uploaded videos statically
	r.Static("/uploads/videos", "./data/videos")

	// health check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
