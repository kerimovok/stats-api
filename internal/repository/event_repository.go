package repository

import (
	"context"
	"stats-api/internal/models"
	"stats-api/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EventRepo provides methods for interacting with the events collection in MongoDB
var EventRepo = &EventRepository{}

type EventRepository struct {
	collection *mongo.Collection
}

// SetCollection configures the MongoDB collection for event operations
func (r *EventRepository) SetCollection(db *mongo.Database, collectionName string) {
	r.collection = db.Collection(collectionName)
}

// QueryEvents retrieves events with pagination, filtering, and sorting
// Parameters:
// - ctx: Context for the operation
// - filters: MongoDB query filters
// - sort: MongoDB sort specification
// - page: Page number (1-based)
// - limit: Maximum number of items per page
func (r *EventRepository) QueryEvents(ctx context.Context, filters bson.M, sort bson.D, page, limit int) ([]models.Event, error) {
	skip, perPage := utils.Pagination(page, limit)

	opts := options.Find().
		SetSort(sort).
		SetSkip(int64(skip)).
		SetLimit(int64(perPage))

	cursor, err := r.collection.Find(ctx, filters, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []models.Event
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

// AggregateStats performs statistical aggregations on events
// Parameters:
// - ctx: Context for the operation
// - filters: MongoDB query filters
// - groupBy: Field to group results by
// - aggregates: Type of aggregation to perform (count, sum, avg)
func (r *EventRepository) AggregateStats(ctx context.Context, filters bson.M, groupBy, aggregates string) ([]bson.M, error) {
	pipeline := []bson.M{}

	// Match stage for filters
	if len(filters) > 0 {
		pipeline = append(pipeline, bson.M{"$match": filters})
	}

	// Group stage
	groupStage := bson.M{
		"_id": "$" + groupBy,
	}

	// Add aggregation operations
	switch aggregates {
	case "count":
		groupStage["value"] = bson.M{"$sum": 1}
	case "sum":
		groupStage["value"] = bson.M{"$sum": "$value"}
	case "avg":
		groupStage["value"] = bson.M{"$avg": "$value"}
	default:
		groupStage["value"] = bson.M{"$sum": 1} // Default to count
	}

	pipeline = append(pipeline,
		bson.M{"$group": groupStage},
		bson.M{"$sort": bson.M{"_id": 1}},
	)

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// AggregateTimeSeries performs time-based aggregations on events
// Parameters:
// - ctx: Context for the operation
// - filters: MongoDB query filters
// - interval: Time interval for grouping (hour, day, week, month)
// - aggregates: Type of aggregation to perform (count, sum, avg)
func (r *EventRepository) AggregateTimeSeries(ctx context.Context, filters bson.M, interval, aggregates string) ([]bson.M, error) {
	pipeline := []bson.M{}

	// Match stage for filters
	if len(filters) > 0 {
		pipeline = append(pipeline, bson.M{"$match": filters})
	}

	// Group by time interval
	groupStage := bson.M{
		"_id": bson.M{
			"$dateToString": bson.M{
				"format": getTimeFormat(interval),
				"date":   "$created_at",
			},
		},
	}

	// Add aggregation operations
	switch aggregates {
	case "count":
		groupStage["value"] = bson.M{"$sum": 1}
	case "sum":
		groupStage["value"] = bson.M{"$sum": "$value"}
	case "avg":
		groupStage["value"] = bson.M{"$avg": "$value"}
	default:
		groupStage["value"] = bson.M{"$sum": 1} // Default to count
	}

	pipeline = append(pipeline,
		bson.M{"$group": groupStage},
		bson.M{"$sort": bson.M{"_id": 1}},
	)

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// getTimeFormat returns the date format string for MongoDB based on the interval
func getTimeFormat(interval string) string {
	switch interval {
	case "hour":
		return "%Y-%m-%d-%H"
	case "day":
		return "%Y-%m-%d"
	case "week":
		return "%Y-%U"
	case "month":
		return "%Y-%m"
	default:
		return "%Y-%m-%d" // Default to daily
	}
}
