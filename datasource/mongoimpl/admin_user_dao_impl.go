package mongoimpl

import (
	"context"
	"dashfun_gamecenter/admin"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math"
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

func (a *AdminUserDaoImpl) SearchUser(name, email string, status admin.AdminUserStatus, size, page int64) (users []*admin.AdminUser, totalPages int, err error) {
	filter := bson.D{}
	if name != "" {
		filter = append(filter, bson.E{
			Key: "name",
			Value: bson.D{
				{"$regex", name},
				{"$options", "i"},
			},
		})
	}

	if email != "" {
		filter = append(filter, bson.E{
			Key: "email",
			Value: bson.D{
				{"$regex", email},
				{"$options", "i"},
			},
		})
	}

	if status > 0 {
		filter = append(filter, bson.E{
			Key: "status",
			Value: bson.D{
				{"$eq", status},
			},
		})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	skip := (page - 1) * size

	totalDocs, err := a.c.CountDocuments(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}

	totalPages = int(math.Ceil(float64(totalDocs) / float64(size)))

	find, err := a.c.Find(ctx, filter, options.Find().SetSkip(skip).SetLimit(size))

	if err != nil {
		return nil, 0, err
	}

	if err = find.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	return users, totalPages, nil
}
