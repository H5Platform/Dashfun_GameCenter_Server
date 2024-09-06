package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type CoinRecordDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetCoinRecordDaoMongo() *CoinRecordDaoMongo {
	dao := &CoinRecordDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (cd *CoinRecordDaoMongo) initDB() {
	c := cd.db.Collection("coin_record_data")
	cd.c = c

	err := CreateIndexes(c, []IndexInfo{
		{
			FieldName: "user_id",
			Unique:    false,
			Sort:      1,
			IndexName: "index_user_id",
		},
		{
			FieldName: "coin_id",
			Unique:    false,
			Sort:      1,
			IndexName: "index_coin_id",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

func (cd *CoinRecordDaoMongo) AddRecord(record *data.CoinUserRecordData) (*data.CoinUserRecordData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := cd.c.InsertOne(ctx, record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (cd *CoinRecordDaoMongo) GetAllUserCoinRecords(userId, coinId string) ([]*data.CoinUserRecordData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ret := make([]*data.CoinUserRecordData, 0)

	opts := options.Find()
	opts.SetSort(bson.M{"time": -1})

	cur, err := cd.c.Find(ctx, bson.M{"user_id": userId, "coin_id": coinId}, opts)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ret, nil
	}
	if err != nil {
		return nil, err
	}

	for cur.Next(ctx) {
		var record data.CoinUserRecordData
		err := cur.Decode(&record)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &record)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}
