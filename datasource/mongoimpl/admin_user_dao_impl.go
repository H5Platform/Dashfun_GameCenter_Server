package mongoimpl

import (
	"context"
	"dashfun_gamecenter/admin"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type AdminUserDaoImpl struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetAdminUserDaoMongo() *AdminUserDaoImpl {
	dao := &AdminUserDaoImpl{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (a *AdminUserDaoImpl) initDB() {
	c := a.db.Collection("admin_user_data")
	a.c = c

	err := CreateIndexes(c, []IndexInfo{
		{
			FieldName: "name",
			Unique:    true,
			Sort:      1,
			IndexName: "index_name",
		},
		{
			FieldName: "email",
			Unique:    true,
			Sort:      1,
			IndexName: "index_email",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

func (a *AdminUserDaoImpl) FindUserById(id string) (*admin.AdminUser, error) {
	var ret *admin.AdminUser
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := a.c.FindOne(ctx, bson.M{"_id": id}).Decode(&ret)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return ret, nil
}

func (a *AdminUserDaoImpl) FindUserByMail(email string) (*admin.AdminUser, error) {
	var ret *admin.AdminUser
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := a.c.FindOne(ctx, bson.M{"email": email}).Decode(&ret)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return ret, nil
}

func (a *AdminUserDaoImpl) FindUserByName(name string) (*admin.AdminUser, error) {
	var ret *admin.AdminUser
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := a.c.FindOne(ctx, bson.M{"name": name}).Decode(&ret)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return ret, nil
}

func (a *AdminUserDaoImpl) SaveUser(user *admin.AdminUser) (*admin.AdminUser, error) {
	update := bson.M{
		"$set": user,
	}
	opts := options.Update().SetUpsert(true)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := a.c.UpdateByID(ctx, user.Id, update, opts)
	if err != nil {
		return nil, err
	}
	return user, nil
}
