package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetPricePredictDaoMongo() *PricePredictDaoMongo {
	dao := &PricePredictDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

type PricePredictDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func (p *PricePredictDaoMongo) SaveOrUpdate(ppData *data.PricePredictData) error {
	update := bson.M{
		"$set": ppData,
	}
	opts := options.Update().SetUpsert(true)
	_, err := p.c.UpdateByID(context.TODO(), ppData.Id, update, opts)
	return err
}

func (p *PricePredictDaoMongo) FindById(id string) (*data.PricePredictData, error) {
	var ret *data.PricePredictData
	err := p.c.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *PricePredictDaoMongo) FindByUserAndDate(userId, date string) (*data.PricePredictData, error) {
	// With single record per user, checking by ID (userId) and verifying date
	ret, err := p.FindById(userId)
	if err != nil {
		return nil, err
	}
	if ret.PredictDate != date {
		return nil, mongo.ErrNoDocuments
	}
	return ret, nil
}

func (p *PricePredictDaoMongo) FindLatestByUser(userId string) (*data.PricePredictData, error) {
	// With single record per user, FindById is sufficient
	return p.FindById(userId)
}

func (p *PricePredictDaoMongo) FindPendingPredictionsByDate(date string) ([]*data.PricePredictData, error) {
	var ret []*data.PricePredictData
	// Find all records for the date that are Unsubmitted or Pending
	filter := bson.M{
		"predict_date": date,
		"status": bson.M{
			"$in": []data.PricePredictStatus{
				data.PricePredictStatusUnsubmitted,
				data.PricePredictStatusPending,
			},
		},
	}
	// Maybe limit? For now, fetch all.
	cursor, err := p.c.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (p *PricePredictDaoMongo) FindBots(ctx context.Context, limit int64) ([]*data.PricePredictData, error) {
	opts := options.Find().SetLimit(limit)
	cursor, err := p.c.Find(ctx, bson.M{"is_bot": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bots []*data.PricePredictData
	if err := cursor.All(ctx, &bots); err != nil {
		return nil, err
	}
	return bots, nil
}

func (p *PricePredictDaoMongo) initDB() {
	c := p.db.Collection("price_predict_data")
	p.c = c

	err := CreateIndexes(c, []IndexInfo{
		{
			FieldName: "predict_date",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_predict_date",
		},
		{
			FieldName: "is_bot",
			Unique:    false,
			Sort:      1,
			IndexName: "index_is_bot",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
