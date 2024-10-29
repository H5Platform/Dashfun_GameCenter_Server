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

type CoinDaoMongo struct {
	db *mongo.Database
	c  *mongo.Collection
}

func GetCoinDaoMongo() *CoinDaoMongo {
	dao := &CoinDaoMongo{
		db: GetMongoDatabase(),
	}
	dao.initDB()
	return dao
}

func (cd *CoinDaoMongo) initDB() {
	c := cd.db.Collection("coin_data")
	cd.c = c

	err := CreateIndexes(c, []IndexInfo{
		{
			FieldName: "name",
			Unique:    true,
			Sort:      1,
			IndexName: "index_name",
		}, {
			FieldName: "bind_game_id",
			Unique:    false,
			Sort:      1,
			IndexName: "index_bind_game_id",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

func (cd *CoinDaoMongo) GetAllCoins() ([]*data.CoinData, error) {
	ret := make([]*data.CoinData, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := cd.c.Find(ctx, bson.M{})
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ret, nil
	}
	if err != nil {
		return nil, err
	}

	for cur.Next(ctx) {
		var c data.CoinData
		err := cur.Decode(&c)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &c)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (cd *CoinDaoMongo) SaveOrUpdate(coin *data.CoinData) (*data.CoinData, error) {
	update := bson.M{
		"$set": coin,
	}
	opts := options.Update().SetUpsert(true)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := cd.c.UpdateByID(ctx, coin.Id, update, opts)
	if err != nil {
		return nil, err
	}

	return coin, nil
}

func (cd *CoinDaoMongo) FindCoinById(coinId string) (*data.CoinData, error) {
	var ret *data.CoinData
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := cd.c.FindOne(ctx, bson.M{"_id": coinId}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (cd *CoinDaoMongo) FindCoinByName(name string) (*data.CoinData, error) {
	var ret *data.CoinData
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := cd.c.FindOne(ctx, bson.M{"name": name}).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// FindCoinByGameId 根据游戏id寻找绑定的coin，找不到返回nil
func (cd *CoinDaoMongo) FindCoinByGameId(gameId string) *data.CoinData {
	var ret *data.CoinData
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := cd.c.FindOne(ctx, bson.M{"bind_game_id": gameId}).Decode(&ret)
	if err != nil {
		return nil
	}
	return ret
}

func (cd *CoinDaoMongo) CreateCoin(id, name, symbol, desc, bindGameId string, canWithdraw bool, minWithdraw float32, chainAddr map[string]string) (*data.CoinData, error) {
	coin := &data.CoinData{
		Id:          id,
		Name:        name,
		Symbol:      symbol,
		Desc:        desc,
		BindGameId:  bindGameId,
		CanWithdraw: canWithdraw,
		MinWithdraw: minWithdraw,
		ChainAddr:   chainAddr,
		CreateTime:  time.Now().UnixMilli(),
	}
	coin, err := cd.SaveOrUpdate(coin)
	if err != nil {
		return nil, err
	}
	return coin, nil
}
