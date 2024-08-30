package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type UserDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetUserDaoMongo() *UserDaoMongo {
	dao := &UserDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (u *UserDaoMongo) initDB() {
	//create user collection
	c := u.db.Collection("user_data")
	u.c = c

	err := CreateIndexes(c, []IndexInfo{
		{
			FieldName: "channel_id",
			Unique:    true,
			Sort:      1,
			IndexName: "index_channel_id",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

func (u *UserDaoMongo) GetUserById(userId string) (*data.DashFunUser, error) {
	var ret *data.DashFunUser
	err := u.c.FindOne(context.TODO(), bson.M{"_id": userId}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (u *UserDaoMongo) GetUserByChannelId(channelId string) (*data.DashFunUser, error) {
	var ret *data.DashFunUser
	err := u.c.FindOne(context.TODO(), bson.M{"channel_id": channelId}).Decode(&ret)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (u *UserDaoMongo) SaveOrUpdate(user *data.DashFunUser) (*data.DashFunUser, error) {
	update := bson.M{
		"$set": user,
	}
	opts := options.Update().SetUpsert(true)
	r, err := u.c.UpdateByID(context.TODO(), user.Id, update, opts)
	log.Printf("r:%+v", r)
	if err != nil {
		return nil, err
	}

	return user, nil
}
