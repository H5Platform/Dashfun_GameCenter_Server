package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type UserProfileDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetUserProfileDaoMongo() *UserProfileDaoMongo {
	dao := &UserProfileDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (u *UserProfileDaoMongo) initDB() {
	c := u.db.Collection("user_profile")
	u.c = c

	err := CreateIndexes(c, []IndexInfo{
		{
			FieldName: "nickname",
			Unique:    true,
			Sort:      -1,
			IndexName: "index_nickname",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

func (u *UserProfileDaoMongo) SaveOrUpdate(profile *data.UserProfileData) (*data.UserProfileData, error) {
	update := bson.M{
		"$set": profile,
	}
	opts := options.Update().SetUpsert(true)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := u.c.UpdateByID(ctx, profile.UserId, update, opts)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (u *UserProfileDaoMongo) GetUserProfile(userId string) (*data.UserProfileData, error) {
	var ret *data.UserProfileData
	err := u.c.FindOne(context.TODO(), bson.M{"_id": userId}).Decode(&ret)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return ret, nil
}
