package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AirdropDaoMongoImpl struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetAirdropDaoMongo() *AirdropDaoMongoImpl {
	dao := &AirdropDaoMongoImpl{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (a *AirdropDaoMongoImpl) initDB() {
	//create user collection
	c := a.db.Collection("airdrop_data")
	a.c = c
}

func (a *AirdropDaoMongoImpl) SaveOrUpdate(airDrop *data.AirdropData) (*data.AirdropData, error) {
	update := bson.M{
		"$set": airDrop,
	}
	opts := options.Update().SetUpsert(true)
	_, err := a.c.UpdateByID(context.TODO(), airDrop.UserId, update, opts)
	if err != nil {
		return nil, err
	}

	return airDrop, nil
}

func (a *AirdropDaoMongoImpl) GetAirdropData(userId string) (*data.AirdropData, error) {
	var ret *data.AirdropData
	err := a.c.FindOne(context.TODO(), bson.M{"_id": userId}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (a *AirdropDaoMongoImpl) GetAllAirdropData() ([]*data.AirdropData, error) {
	cursor, err := a.c.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var items []*data.AirdropData
	if err = cursor.All(context.TODO(), &items); err != nil {
		return nil, err
	}

	if err = cursor.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
