package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type FishingLeaderboardBotDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetFishingLeaderboardBotDaoMongo() *FishingLeaderboardBotDaoMongo {
	dao := &FishingLeaderboardBotDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (l *FishingLeaderboardBotDaoMongo) initDB() {
	c := l.db.Collection("fishing_leaderboard_bot_data")
	l.c = c
}

func (l *FishingLeaderboardBotDaoMongo) SaveOrUpdate(bot *data.FishingBotData) (*data.FishingBotData, error) {
	update := bson.M{
		"$set": bot,
	}
	opts := options.Update().SetUpsert(true)
	_, err := l.c.UpdateByID(context.TODO(), bot.Id, update, opts)
	if err != nil {
		return nil, err
	}

	return bot, nil
}

func (l *FishingLeaderboardBotDaoMongo) LoadAllBots() ([]*data.FishingBotData, error) {
	var ret []*data.FishingBotData
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := l.c.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &ret)
	return ret, nil
}
