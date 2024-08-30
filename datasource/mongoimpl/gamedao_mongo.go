package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type GameDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func (g *GameDaoMongo) GetGameById(gameId string) (*data.DashFunGame, error) {
	var ret *data.DashFunGame
	err := g.c.FindOne(context.TODO(), bson.M{"_id": gameId}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (g *GameDaoMongo) SaveOrUpdate(game *data.DashFunGame) (*data.DashFunGame, error) {
	update := bson.M{
		"$set": game,
	}
	opts := options.Update().SetUpsert(true)
	_, err := g.c.UpdateByID(context.TODO(), game.Id, update, opts)
	if err != nil {
		return nil, err
	}

	return game, nil
}

func GetGameDaoMongo() *GameDaoMongo {
	dao := &GameDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (g *GameDaoMongo) initDB() {
	c := g.db.Collection("game_data")
	g.c = c

	err := CreateIndexes(c, []IndexInfo{
		{
			FieldName: "time",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_time",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
