package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/datasource/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ExchangeDaoMongo struct {
	c *mongo.Collection
}

func NewExchangeDaoMongo(db *mongo.Database) types.ExchangeDao {
	return &ExchangeDaoMongo{
		c: db.Collection(data.CollectionExchangeLog),
	}
}

func (d *ExchangeDaoMongo) Save(ctx context.Context, log *data.ExchangeLog) error {
	if log.Id == "" {
		log.Id = primitive.NewObjectID().Hex()
	}
	_, err := d.c.InsertOne(ctx, log)
	return err
}

func (d *ExchangeDaoMongo) UpdateStatus(ctx context.Context, id string, status int, txHash string) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": status, "tx_hash": txHash}}
	_, err := d.c.UpdateOne(ctx, filter, update)
	return err
}

func (d *ExchangeDaoMongo) GetUnissuedLogs(ctx context.Context) ([]*data.ExchangeLog, error) {
	filter := bson.M{"status": data.ExchangeStatusUnissued}
	// Sort by create_time asc to process older requests first
	opts := options.Find().SetSort(bson.M{"create_time": 1})

	cursor, err := d.c.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*data.ExchangeLog
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (d *ExchangeDaoMongo) GetExchangeHistory(ctx context.Context, userId string, limit int) ([]*data.ExchangeLog, error) {
	filter := bson.M{"user_id": userId}
	// Sort by create_time desc
	opts := options.Find().SetSort(bson.M{"create_time": -1}).SetLimit(int64(limit))

	cursor, err := d.c.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*data.ExchangeLog
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (d *ExchangeDaoMongo) GetDailyGlobalUsage(ctx context.Context, date string) (float64, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"date": date}},
		{"$group": bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$token_amount"},
		}},
	}

	cursor, err := d.c.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) > 0 {
		// total might be int or float, handle safely
		// In Go mongo driver, usually it decodes to appropriate type.
		// Since TokenAmount is float64, sum should be float64 (or int32/int64 if integers)
		// Let's use flexible helper or type assertion
		if val, ok := results[0]["total"]; ok {
			return toFloat64(val), nil
		}
	}
	return 0, nil
}

func (d *ExchangeDaoMongo) GetDailyUserUsage(ctx context.Context, userId, date string) (float64, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userId, "date": date}},
		{"$group": bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$token_amount"},
		}},
	}

	cursor, err := d.c.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) > 0 {
		if val, ok := results[0]["total"]; ok {
			return toFloat64(val), nil
		}
	}
	return 0, nil
}

func toFloat64(v interface{}) float64 {
	switch i := v.(type) {
	case float64:
		return i
	case float32:
		return float64(i)
	case int64:
		return float64(i)
	case int32:
		return float64(i)
	case int:
		return float64(i)
	default:
		return 0
	}
}
