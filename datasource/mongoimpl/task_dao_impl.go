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

type TaskDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetTaskDaoMongo() *TaskDaoMongo {
	dao := &TaskDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (p *TaskDaoMongo) initDB() {
	c := p.db.Collection("task_data")
	p.c = c

	err := CreateIndexes(c, []IndexInfo{
		{
			FieldName: "game_id",
			Unique:    false,
			Sort:      1,
			IndexName: "index_game_id",
		},
		{
			FieldName: "name",
			Unique:    true,
			Sort:      1,
			IndexName: "index_name",
		},
		{
			FieldName: "task_type",
			Unique:    false,
			Sort:      1,
			IndexName: "index_task_type",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

func (p *TaskDaoMongo) FindAllTasks() []*data.DashFunTaskData {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var ret []*data.DashFunTaskData

	cur, err := p.c.Find(ctx, bson.M{})
	if errors.Is(err, mongo.ErrNoDocuments) {
		return make([]*data.DashFunTaskData, 0)
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var d data.DashFunTaskData
		err := cur.Decode(&d)

		if err != nil {
			return make([]*data.DashFunTaskData, 0)
		}
		ret = append(ret, &d)
	}

	if err := cur.Err(); err != nil {
		return make([]*data.DashFunTaskData, 0)
	}
	return ret
}

func (p *TaskDaoMongo) FindTaskById(id string) (*data.DashFunTaskData, error) {
	var ret *data.DashFunTaskData
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := p.c.FindOne(ctx, bson.M{"_id": id}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *TaskDaoMongo) FindTaskByName(name string) (*data.DashFunTaskData, error) {
	var ret *data.DashFunTaskData
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := p.c.FindOne(ctx, bson.M{"name": name}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *TaskDaoMongo) SaveOrUpdate(task *data.DashFunTaskData) (*data.DashFunTaskData, error) {
	update := bson.M{
		"$set": task,
	}
	opts := options.Update().SetUpsert(true)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := p.c.UpdateByID(ctx, task.Id, update, opts)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (p *TaskDaoMongo) CreateTask(id, name, gameId string, taskType data.DashFunTaskType, category data.DashFunTaskCategory,
	condition data.DashFunTaskCondition, reward data.DashFunTaskReward) (*data.DashFunTaskData, error) {
	task := &data.DashFunTaskData{
		Id:         id,
		Name:       name,
		Open:       true,
		GameId:     gameId,
		Type:       taskType,
		Category:   category,
		Condition:  condition,
		Reward:     reward,
		CreateTime: time.Now().UnixMilli(),
	}
	task, err := p.SaveOrUpdate(task)
	if err != nil {
		return nil, err
	}
	return task, nil
}
