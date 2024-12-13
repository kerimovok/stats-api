package handlers

import (
	"context"
	"strconv"
	"time"

	"stats-api/internal/repository"
	internalUtils "stats-api/internal/utils"
	pkgUtils "stats-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// GetEvents retrieves a paginated list of events with optional filtering and sorting
// Supports query parameters:
// - page: Page number (default: 1)
// - limit: Items per page (default: 10)
// - sortBy: Field to sort by (default: createdAt)
// - sortOrder: Sort direction, 'asc' or 'desc' (default: asc)
// - Any other query parameter will be used as a filter
func GetEvents(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Extract query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	sortBy := c.Query("sortBy", "createdAt")
	sortOrder := c.Query("sortOrder", "asc")

	// Extract filters
	filters := bson.M{}
	for key, values := range c.Queries() {
		if key != "page" && key != "limit" && key != "sortBy" && key != "sortOrder" {
			filters[key] = values
		}
	}

	// Query events
	events, err := repository.EventRepo.QueryEvents(ctx, filters, internalUtils.BuildSortOptions(sortBy, sortOrder), page, limit)
	if err != nil {
		return pkgUtils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch events", err)
	}

	return pkgUtils.SuccessResponse(c, "Events retrieved successfully", fiber.Map{
		"page":   page,
		"limit":  limit,
		"events": events,
	})
}

// GetStats aggregates event data based on grouping and aggregation criteria
// Supports query parameters:
// - groupBy: Field to group results by
// - aggregates: Aggregation operation (count, sum, avg)
// - Any other query parameter will be used as a filter
func GetStats(c *fiber.Ctx) error {
	ctx := context.Background()

	// Extract query parameters
	groupBy := c.Query("groupBy", "")
	aggregates := c.Query("aggregates", "count") // E.g., count, sum, avg
	filters := bson.M{}
	for key, values := range c.Queries() {
		if key != "groupBy" && key != "aggregates" {
			filters[key] = values
		}
	}

	// Perform aggregation query
	stats, err := repository.EventRepo.AggregateStats(ctx, filters, groupBy, aggregates)
	if err != nil {
		return pkgUtils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch stats", err)
	}

	return pkgUtils.SuccessResponse(c, "Stats retrieved successfully", fiber.Map{
		"groupBy":    groupBy,
		"aggregates": aggregates,
		"stats":      stats,
	})
}

// GetTimeSeries generates time-based aggregations of event data
// Supports query parameters:
// - interval: Time grouping interval (hour, day, week, month)
// - aggregates: Aggregation operation (count, sum, avg)
// - Any other query parameter will be used as a filter
func GetTimeSeries(c *fiber.Ctx) error {
	ctx := context.Background()

	// Extract query parameters
	aggregates := c.Query("aggregates", "count")
	interval := c.Query("interval", "day") // E.g., day, week, month
	filters := bson.M{}
	for key, values := range c.Queries() {
		if key != "aggregates" && key != "interval" {
			filters[key] = values
		}
	}

	// Perform time-series query
	timeSeries, err := repository.EventRepo.AggregateTimeSeries(ctx, filters, interval, aggregates)
	if err != nil {
		return pkgUtils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch time series", err)
	}

	return pkgUtils.SuccessResponse(c, "Time series retrieved successfully", fiber.Map{
		"interval":   interval,
		"aggregates": aggregates,
		"timeSeries": timeSeries,
	})
}
