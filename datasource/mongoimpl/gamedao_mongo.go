package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math"
	"time"
)

type GameDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func (g *GameDaoMongo) FindGameList(listType data.GameListType, count int) (games []*data.DashFunGame, err error) {
	filter := bson.D{}
	sort := bson.D{{"time", -1}, {"_id", 1}}

	if listType == data.GameListType_New {
		filter = append(filter, bson.E{Key: "status", Value: 2})
		sort = bson.D{{"new_flag", -1}, {"time", -1}, {"_id", 1}}
	} else if listType == data.GameListType_Popular {
		sort = bson.D{{"time", -1}, {"_id", 1}}
		filter = append(filter, bson.E{Key: "popular_flag", Value: 1}, bson.E{Key: "status", Value: 2})
	} else if listType == data.GameListType_Suggest {
		sort = bson.D{{"time", -1}, {"_id", 1}}
		filter = append(filter, bson.E{Key: "suggest_flag", Value: 1})
	} else if listType == data.GameListType_Banner {
		sort = bson.D{{"time", -1}, {"_id", 1}}
		filter = append(filter, bson.E{Key: "banner_flag", Value: 1})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	find, err := g.c.Find(ctx, filter, options.Find().SetLimit(int64(count)).SetSort(sort))

	if err != nil {
		return nil, err
	}

	if err = find.All(ctx, &games); err != nil {
		return nil, err
	}

	return games, nil
}

func (g *GameDaoMongo) GetGameById(gameId string) (*data.DashFunGame, error) {
	var ret *data.DashFunGame
	err := g.c.FindOne(context.TODO(), bson.M{"_id": gameId}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (g *GameDaoMongo) GetGameByName(gameName string) (*data.DashFunGame, error) {
	var ret *data.DashFunGame
	err := g.c.FindOne(context.TODO(), bson.M{"name": gameName}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (g *GameDaoMongo) FindGames(keyword string, genre []int, status data.DashFunGameStatus, size, page int64) (games []*data.DashFunGame, totalPages int, err error) {
	filter := bson.D{}
	if keyword != "" {
		filter = append(filter, bson.E{
			Key: "name",
			Value: bson.D{
				{"$regex", keyword},
				{"$options", "i"},
			},
		})
	}

	if genre != nil && len(genre) > 0 {
		filter = append(filter, bson.E{
			Key: "genre",
			Value: bson.D{
				{"$all", bson.A{genre}},
			},
		})
	}

	if status > data.DashFunGameStatus_NoChange {
		filter = append(filter, bson.E{
			Key: "status",
			Value: bson.D{
				{"$eq", status},
			},
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	skip := (page - 1) * size

	totalDocs, err := g.c.CountDocuments(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}

	totalPages = int(math.Ceil(float64(totalDocs) / float64(size)))

	find, err := g.c.Find(ctx, filter, options.Find().SetSkip(skip).SetLimit(size).SetSort(bson.D{{"time", -1}}))
	if err != nil {
		return nil, 0, err
	}

	if err = find.All(ctx, &games); err != nil {
		return nil, 0, err
	}

	return games, totalPages, nil
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
			FieldName: "name",
			Unique:    true,
			Sort:      -1,
			IndexName: "index_name",
		},
		{
			FieldName: "time",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_time",
		},
		{
			FieldName: "genre",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_genre",
		},
		{
			FieldName: "status",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_status",
		},
		{
			FieldName: "suggest_flag",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_suggest_flag",
		},
		{
			FieldName: "new_flag",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_new_flag",
		},
		{
			FieldName: "popular_flag",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_popular_flag",
		},
		{
			FieldName: "banner_flag",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_banner_flag",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
