package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DashFunUserPlayRecordDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetUserPlayRecordDaoMongo() *DashFunUserPlayRecordDaoMongo {
	dao := &DashFunUserPlayRecordDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (p *DashFunUserPlayRecordDaoMongo) initDB() {
	c := p.db.Collection("user_play_record")
	p.c = c
}

func (p *DashFunUserPlayRecordDaoMongo) SaveOrUpdate(record *data.DashFunUserPlayRecord) (*data.DashFunUserPlayRecord, error) {
	update := bson.M{
		"$set": record,
	}
	opts := options.Update().SetUpsert(true)
	_, err := p.c.UpdateByID(context.TODO(), record.UserId, update, opts)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (p *DashFunUserPlayRecordDaoMongo) GetUserPlayRecord(userId string) (*data.DashFunUserPlayRecord, error) {
	var ret *data.DashFunUserPlayRecord
	err := p.c.FindOne(context.TODO(), bson.M{"_id": userId}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
