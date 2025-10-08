package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type NolanDevPostDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetNolanDevPostDaoMongo() *NolanDevPostDaoMongo {
	dao := &NolanDevPostDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (u *NolanDevPostDaoMongo) initDB() {
	c := u.db.Collection("nolan_dev_post")
	u.c = c

	err := CreateIndexes(c, []IndexInfo{
		{
			FieldName: "user_id",
			Unique:    false,
			Sort:      1,
			IndexName: "index_user_id",
		},
		{
			FieldName: "created_at",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_created_at",
		}, {
			FieldName: "location",
			Unique:    false,
			Sort:      1,
			IndexName: "index_location",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

func (u *NolanDevPostDaoMongo) SaveOrUpdate(post *data.NolanPostData) (*data.NolanPostData, error) {
	update := bson.M{
		"$set": post,
	}
	opts := options.Update().SetUpsert(true)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := u.c.UpdateByID(ctx, post.PostId, update, opts)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (u *NolanDevPostDaoMongo) GetPosts(limit int) ([]*data.NolanPostData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{"created_at", -1}})

	cursor, err := u.c.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []*data.NolanPostData
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (u *NolanDevPostDaoMongo) GetUserLatestPostTime(userId string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userId}
	findOptions := options.FindOne().SetSort(bson.D{{"created_at", -1}})

	var post data.FishingPostData
	err := u.c.FindOne(ctx, filter, findOptions).Decode(&post)
	if err != nil {
		return 0, err
	}

	return post.CreatedAt, nil
}
