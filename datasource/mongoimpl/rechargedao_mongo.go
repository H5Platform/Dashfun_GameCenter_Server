package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func GetRechargeDaoMongo() *RechargeDaoMongo {
	dao := &RechargeDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

type RechargeDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func (p *RechargeDaoMongo) FindRechargeById(id string) (*data.DashFunRechargeData, error) {
	var ret *data.DashFunRechargeData
	err := p.c.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *RechargeDaoMongo) GetOrdersByStatus(status data.RechargeStatus) ([]*data.DashFunRechargeData, error) {
	var ret []*data.DashFunRechargeData
	cursor, err := p.c.Find(context.TODO(), bson.M{"status": status})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	err = cursor.All(context.TODO(), &ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (p *RechargeDaoMongo) CreateRecharge(id, userId string, from data.RechargeFrom, gameId string, price int, priceType data.RechargePlatformOptionPriceType, diamond int,
	payload, message string, createAt int64) (*data.DashFunRechargeData, error) {
	recharge := &data.DashFunRechargeData{
		Id:           id,
		UserId:       userId,
		ChannelPayId: "",
		GameId:       gameId,
		From:         from,
		Price:        price,
		PriceType:    priceType,
		Diamond:      diamond,
		Payload:      payload,
		Message:      message,
		CreatedAt:    createAt,
		PaidAt:       0,
		Status:       data.DashFunRechargeStatus_Created,
	}
	return p.SaveOrUpdate(recharge)
}

func (p *RechargeDaoMongo) SaveOrUpdate(recharge *data.DashFunRechargeData) (*data.DashFunRechargeData, error) {
	update := bson.M{
		"$set": recharge,
	}
	opts := options.Update().SetUpsert(true)
	_, err := p.c.UpdateByID(context.TODO(), recharge.Id, update, opts)
	if err != nil {
		return nil, err
	}

	return recharge, nil
}

func (p *RechargeDaoMongo) initDB() {
	c := p.db.Collection("recharge_data")
	p.c = c

	err := CreateIndexes(c, []IndexInfo{
		{
			FieldName: "user_id",
			Unique:    false,
			Sort:      1,
			IndexName: "index_user_id",
		},
		{
			FieldName: "price_type",
			Unique:    false,
			Sort:      1,
			IndexName: "index_price_type",
		},
		{
			FieldName: "game_id",
			Unique:    false,
			Sort:      1,
			IndexName: "index_game_id",
		},
		{
			FieldName: "create_at",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_create_at",
		},
		{
			FieldName: "from",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_from",
		},
		{
			FieldName: "status",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_status",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
