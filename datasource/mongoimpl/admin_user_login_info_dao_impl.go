package mongoimpl

import (
	"context"
	"dashfun_gamecenter/admin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type AdminUserLoginInfoDaoImpl struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetAdminUserLoginInfoDaoMongo() *AdminUserLoginInfoDaoImpl {
	dao := &AdminUserLoginInfoDaoImpl{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (a *AdminUserLoginInfoDaoImpl) initDB() {
	c := a.db.Collection("admin_user_login_info_data")
	a.c = c
}

func (a *AdminUserLoginInfoDaoImpl) FindUserLoginInfo(id string) (*admin.AdminUserLoginInfo, error) {
	var ret *admin.AdminUserLoginInfo
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := a.c.FindOne(ctx, bson.M{"_id": id}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (a *AdminUserLoginInfoDaoImpl) SaveUserLoginInfo(info *admin.AdminUserLoginInfo) (*admin.AdminUserLoginInfo, error) {
	update := bson.M{
		"$set": info,
	}
	opts := options.Update().SetUpsert(true)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := a.c.UpdateByID(ctx, info.Id, update, opts)
	if err != nil {
		return nil, err
	}
	return info, nil
}
