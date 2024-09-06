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

type CoinUserDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetCoinUserDaoMongo() *CoinUserDaoMongo {
	dao := &CoinUserDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (cd *CoinUserDaoMongo) initDB() {
	c := cd.db.Collection("coin_user_data")
	cd.c = c

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"user_id", 1}, {"coin_id", 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := c.Indexes().CreateOne(context.TODO(), indexModel)

	if err != nil {
		log.Fatal(err)
	}
}

func (cd *CoinUserDaoMongo) SaveOrUpdate(userData *data.CoinUserData) (*data.CoinUserData, error) {
	update := bson.M{
		"$set": userData,
	}
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"user_id": userData.UserId, "coin_id": userData.CoinId}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := cd.c.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}

	return userData, nil
}

func (cd *CoinUserDaoMongo) GetAllUserCoins(userId string) ([]*data.CoinUserData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ret := make([]*data.CoinUserData, 0)
	cur, err := cd.c.Find(ctx, bson.M{"user_id": userId})
	defer cur.Close(ctx)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return ret, nil
	}
	if err != nil {
		return nil, err
	}

	for cur.Next(ctx) {
		var d data.CoinUserData
		err := cur.Decode(&d)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &d)
	}

	if err = cur.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}
