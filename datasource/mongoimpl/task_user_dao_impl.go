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

type TaskUserDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetTaskUserDaoMongo() *TaskUserDaoMongo {
	dao := &TaskUserDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (p *TaskUserDaoMongo) initDB() {
	c := p.db.Collection("task_user_data")
	p.c = c

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"user_id", 1}, {"task_id", 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := c.Indexes().CreateOne(context.TODO(), indexModel)

	if err != nil {
		log.Fatal(err)
	}

	err = CreateIndexes(c, []IndexInfo{
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

func (p *TaskUserDaoMongo) FindTaskUserData(taskId string, userId string) (*data.DashFunTaskUserData, error) {
	var ret *data.DashFunTaskUserData
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := p.c.FindOne(ctx, bson.M{"user_id": userId, "task_id": taskId}).Decode(&ret)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *TaskUserDaoMongo) FindAllTaskUserData(userId string) ([]*data.DashFunTaskUserData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ret := make([]*data.DashFunTaskUserData, 0)
	cursor, err := p.c.Find(ctx, bson.M{"user_id": userId})
	defer cursor.Close(ctx)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ret, nil
	}
	if err != nil {
		return nil, err
	}

	for cursor.Next(ctx) {
		var d data.DashFunTaskUserData
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

func (p *TaskUserDaoMongo) SaveOrUpdate(userData *data.DashFunTaskUserData) (*data.DashFunTaskUserData, error) {
	update := bson.M{
		"$set": userData,
	}
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"user_id": userData.UserId, "task_id": userData.TaskId}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := p.c.UpdateOne(ctx, filter, update, opts)
	//_, err := p.c.UpdateByID(context.TODO(), userData.UserId, update, opts)
	if err != nil {
		return nil, err
	}

	return userData, nil
}
