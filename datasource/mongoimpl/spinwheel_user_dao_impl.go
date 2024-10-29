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

type SpinWheelUserDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetSpinWheelUserDaoMongo() *SpinWheelUserDaoMongo {
	dao := &SpinWheelUserDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (s *SpinWheelUserDaoMongo) initDB() {
	c := s.db.Collection("spinwheel_user_data")
	s.c = c

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"user_id", 1}, {"spin_wheel_id", 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := c.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Fatal(err)
	}

	err = CreateIndexes(c, []IndexInfo{
		{
			FieldName: "status",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_status",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

func (s *SpinWheelUserDaoMongo) SaveOrUpdate(userData *data.SpinWheelUserData) (*data.SpinWheelUserData, error) {
	update := bson.M{
		"$set": userData,
	}
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"user_id": userData.UserId, "spin_wheel_id": userData.SpinWheelId}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := s.c.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}

	return userData, nil
}

func (s *SpinWheelUserDaoMongo) GetUserSpinWheelData(userId, spinWheelId string) (*data.SpinWheelUserData, error) {
	var ret *data.SpinWheelUserData
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.c.FindOne(ctx, bson.M{"user_id": userId, "spin_wheel_id": spinWheelId}).Decode(&ret)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ret, nil
}
