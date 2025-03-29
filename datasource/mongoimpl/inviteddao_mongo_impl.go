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

type InvitedUserDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetInvitedUserDaoMongo() *InvitedUserDaoMongo {
	dao := &InvitedUserDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (i *InvitedUserDaoMongo) initDB() {
	//create user collection
	c := i.db.Collection("invited_data")
	i.c = c

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"user_id", 1}, {"invited_user_id", 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := c.Indexes().CreateOne(context.TODO(), indexModel)

	if err != nil {
		log.Fatal(err)
	}

	err = CreateIndexes(c, []IndexInfo{
		{
			FieldName: "invited_status",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_invited_status",
		},
		{
			FieldName: "invited_time",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_invited_time",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

func (i *InvitedUserDaoMongo) FindInvitedByUserId(userId string) ([]*data.InvitedUserData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ret := make([]*data.InvitedUserData, 0)
	cursor, err := i.c.Find(ctx, bson.M{"user_id": userId}, options.Find().SetSort(bson.D{{Key: "invited_time", Value: -1}}))
	defer cursor.Close(ctx)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ret, nil
	}
	if err != nil {
		return nil, err
	}

	for cursor.Next(ctx) {
		var d data.InvitedUserData
		err := cursor.Decode(&d)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &d)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

// FindInvitedByInvitedUserId 根据被邀请人ID查找邀请人
// 一个用户可能被多个用户邀请，所以返回的是一个数组
func (i *InvitedUserDaoMongo) FindInvitedByInvitedUserId(userId string) ([]*data.InvitedUserData, error) {
	cursor, err := i.c.Find(context.TODO(), bson.M{"invited_user_id": userId})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return make([]*data.InvitedUserData, 0), nil
		}
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var ret []*data.InvitedUserData
	err = cursor.All(context.TODO(), &ret)
	if err != nil {
		return nil, err
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (i *InvitedUserDaoMongo) SaveOrUpdate(userData *data.InvitedUserData) (*data.InvitedUserData, error) {
	update := bson.M{
		"$set": userData,
	}
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"user_id": userData.UserId, "invited_user_id": userData.InvitedUserId}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := i.c.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}

	return userData, nil
}
