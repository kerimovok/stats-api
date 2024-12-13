package routes

import (
	"stats-api/internal/handlers"
	"stats-api/internal/repository"
	"stats-api/pkg/database"
	"stats-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
)

func Setup(app *fiber.App) {
	// Initialize repository collection
	db := database.DBClient.Database(utils.GetEnv("DB_NAME"))
	repository.EventRepo.SetCollection(db, "events")

	// API routes group
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Monitor route
	app.Get("/metrics", monitor.New())

	// Event routes
	event := v1.Group("/events")
	event.Get("/", handlers.GetEvents)
	event.Get("/stats", handlers.GetStats)
	event.Get("/timeseries", handlers.GetTimeSeries)
}
