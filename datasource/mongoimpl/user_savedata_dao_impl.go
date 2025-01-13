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

type UserSaveDataDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetUserSaveDataDaoMongo() *UserSaveDataDaoMongo {
	dao := &UserSaveDataDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (u *UserSaveDataDaoMongo) initDB() {
	c := u.db.Collection("user_savedata")
	u.c = c

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"user_id", 1}, {"game_id", 1}, {"key", 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := c.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Fatal(err)
	}
}

func (u *UserSaveDataDaoMongo) SaveOrUpdate(saveData *data.DashFunUserSaveData) (*data.DashFunUserSaveData, error) {
	update := bson.M{
		"$set": saveData,
	}
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"user_id": saveData.UserId, "game_id": saveData.GameId, "key": saveData.Key}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := u.c.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}

	return saveData, nil
}

func (u *UserSaveDataDaoMongo) GetUserSaveData(userId, gameId, key string) (*data.DashFunUserSaveData, error) {
	var ret *data.DashFunUserSaveData
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"user_id": userId, "game_id": gameId, "key": key}
	err := u.c.FindOne(ctx, filter).Decode(&ret)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ret, nil
}
