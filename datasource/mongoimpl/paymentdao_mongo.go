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

func GetPaymentDaoMongo() *PaymentDaoMongo {
	dao := &PaymentDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

type PaymentDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func (p *PaymentDaoMongo) FindPaymentById(id string) (*data.DashFunPaymentData, error) {
	var ret *data.DashFunPaymentData
	err := p.c.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *PaymentDaoMongo) CreatePayment(id, userId, gameId, paymentId, title, desc, payload, currency string, from data.PaymentFrom, price int, extraData string) (*data.DashFunPaymentData, error) {
	payment := &data.DashFunPaymentData{
		Id:          id,
		UserId:      userId,
		GameId:      gameId,
		PaymentId:   paymentId,
		Title:       title,
		Description: desc,
		Payload:     payload,
		Currency:    currency,
		From:        from,
		Price:       price,
		ExtraData:   extraData,
		Message:     "",
		CreatedAt:   time.Now().UnixMilli(),
		Status:      data.DashFunPaymentStatus_Created,
	}
	return p.SaveOrUpdate(payment)
}

func (p *PaymentDaoMongo) SaveOrUpdate(payment *data.DashFunPaymentData) (*data.DashFunPaymentData, error) {
	update := bson.M{
		"$set": payment,
	}
	opts := options.Update().SetUpsert(true)
	_, err := p.c.UpdateByID(context.TODO(), payment.Id, update, opts)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (p *PaymentDaoMongo) initDB() {
	c := p.db.Collection("payment_data")
	p.c = c

	err := CreateIndexes(c, []IndexInfo{
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
			FieldName: "currency",
			Unique:    false,
			Sort:      -1,
			IndexName: "index_create_at",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
