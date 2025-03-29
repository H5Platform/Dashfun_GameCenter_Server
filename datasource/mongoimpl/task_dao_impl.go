package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math"
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
			log.Printf("decode task %s error: %v\n", d.Id, err)
			continue
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

func (p *TaskDaoMongo) FindTaskByGameId(gameId string) ([]*data.DashFunTaskData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var ret []*data.DashFunTaskData

	cur, err := p.c.Find(ctx, bson.M{"game_id": gameId})
	if errors.Is(err, mongo.ErrNoDocuments) {
		return make([]*data.DashFunTaskData, 0), nil
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var d data.DashFunTaskData
		err := cur.Decode(&d)

		if err != nil {
			return make([]*data.DashFunTaskData, 0), err
		}
		ret = append(ret, &d)
	}

	if err := cur.Err(); err != nil {
		return make([]*data.DashFunTaskData, 0), err
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

func (p *TaskDaoMongo) SearchTask(gameId, keyword string, size, page int64) (tasks []*data.DashFunTaskData, totalPages int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.D{}
	//filter = bson.D{
	//	{
	//		"$or", bson.A{
	//			bson.D{{"name", bson.D{{"$regex", keyword}, {"$options", "i"}}}},
	//			bson.D{{"game_id", keyword}},
	//		}},
	//}

	if gameId != "" {
		filter = append(filter, bson.E{Key: "game_id", Value: gameId})
	}
	if keyword != "" {
		filter = append(filter, bson.E{
			Key: "name",
			Value: bson.D{
				{"$regex", keyword},
				{"$options", "i"},
			},
		})
	}

	skip := (page - 1) * size

	totalDocs, err := p.c.CountDocuments(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}

	totalPages = int(math.Ceil(float64(totalDocs) / float64(size)))

	find, err := p.c.Find(ctx, filter, options.Find().SetSkip(skip).SetLimit(size))

	if err != nil {
		return nil, 0, err
	}

	if err = find.All(ctx, &tasks); err != nil {
		return nil, 0, err
	}

	return tasks, totalPages, nil
}

func (p *TaskDaoMongo) CreateTask(id, name, gameId string, taskType data.DashFunTaskType, category data.DashFunTaskCategory,
	condition data.DashFunTaskCondition, rewards []data.DashFunTaskReward) (*data.DashFunTaskData, error) {
	task := &data.DashFunTaskData{
		Id:         id,
		Name:       name,
		Open:       true,
		GameId:     gameId,
		Type:       taskType,
		Category:   category,
		Condition:  condition,
		Rewards:    rewards,
		CreateTime: time.Now().UnixMilli(),
	}
	task, err := p.SaveOrUpdate(task)
	if err != nil {
		return nil, err
	}
	return task, nil
}
