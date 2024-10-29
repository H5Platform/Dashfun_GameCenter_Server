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

type SpinWheelDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetSpinWheelDaoMongo() *SpinWheelDaoMongo {
	dao := &SpinWheelDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (s *SpinWheelDaoMongo) initDB() {
	c := s.db.Collection("spinwheel_data")
	s.c = c
	err := CreateIndexes(c, []IndexInfo{
		{
			FieldName: "name",
			Unique:    true,
			Sort:      -1,
			IndexName: "index_name",
		},
		{
			FieldName: "game_id",
			Unique:    true,
			Sort:      -1,
			IndexName: "index_game_id",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

func (s *SpinWheelDaoMongo) CreateSpinWheel(id, name, gameId string, rewards []data.SpinWheelReward) (*data.SpinWheelData, error) {
	spinWheelData := &data.SpinWheelData{
		Id:      id,
		Name:    name,
		GameId:  gameId,
		Rewards: rewards,
	}
	return s.SaveOrUpdate(spinWheelData)
}

func (s *SpinWheelDaoMongo) GetGameSpinWheel(gameId string) (*data.SpinWheelData, error) {
	var ret *data.SpinWheelData
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.c.FindOne(ctx, bson.M{"game_id": gameId}).Decode(&ret)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *SpinWheelDaoMongo) GetSpinWheelById(spinWheelId string) (*data.SpinWheelData, error) {
	var ret *data.SpinWheelData
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.c.FindOne(ctx, bson.M{"id": spinWheelId}).Decode(&ret)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *SpinWheelDaoMongo) SaveOrUpdate(data *data.SpinWheelData) (*data.SpinWheelData, error) {
	update := bson.M{
		"$set": data,
	}
	opts := options.Update().SetUpsert(true)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := s.c.UpdateByID(ctx, data.Id, update, opts)
	if err != nil {
		return nil, err
	}

	return data, nil
}
