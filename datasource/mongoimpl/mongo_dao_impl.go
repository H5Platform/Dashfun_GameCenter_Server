package mongoimpl

import (
	"context"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var dbDashFun *mongo.Database

// DaoImplMongo implements types.DaoImpl
type DaoImplMongo struct {
	userDao               types.UserDao
	gameDao               types.GameDao
	paymentDao            types.PaymentDao
	taskDao               types.TaskDao
	taskUserDao           types.TaskUserDao
	coinDao               types.CoinDao
	coinUserDao           types.CoinUserDao
	coinRecordDao         types.CoinRecordDao
	adminUserDao          types.AdminUserDao
	adminUserLoginInfoDao types.AdminUserLoginInfoDao
}

func NewDaoImplMongo() *DaoImplMongo {
	return &DaoImplMongo{
		userDao:               GetUserDaoMongo(),
		gameDao:               GetGameDaoMongo(),
		paymentDao:            GetPaymentDaoMongo(),
		taskDao:               GetTaskDaoMongo(),
		taskUserDao:           GetTaskUserDaoMongo(),
		coinDao:               GetCoinDaoMongo(),
		coinUserDao:           GetCoinUserDaoMongo(),
		coinRecordDao:         GetCoinRecordDaoMongo(),
		adminUserDao:          GetAdminUserDaoMongo(),
		adminUserLoginInfoDao: GetAdminUserLoginInfoDaoMongo(),
	}
}

func (d *DaoImplMongo) GetUserDao() types.UserDao {
	return d.userDao
}
func (d *DaoImplMongo) GetGameDao() types.GameDao             { return d.gameDao }
func (d *DaoImplMongo) GetPaymentDao() types.PaymentDao       { return d.paymentDao }
func (d *DaoImplMongo) GetTaskDao() types.TaskDao             { return d.taskDao }
func (d *DaoImplMongo) GetTaskUserDao() types.TaskUserDao     { return d.taskUserDao }
func (d *DaoImplMongo) GetCoinDao() types.CoinDao             { return d.coinDao }
func (d *DaoImplMongo) GetCoinUserDao() types.CoinUserDao     { return d.coinUserDao }
func (d *DaoImplMongo) GetCoinRecordDao() types.CoinRecordDao { return d.coinRecordDao }
func (d *DaoImplMongo) GetAdminUserDao() types.AdminUserDao {
	return d.adminUserDao
}
func (d *DaoImplMongo) GetAdminUserLoginInfoDao() types.AdminUserLoginInfoDao {
	return d.adminUserLoginInfoDao
}

func GetMongoDatabase() *mongo.Database {
	mongoCfg := config.GetConfig().Mongo
	if client == nil {
		serverApi := options.ServerAPI(options.ServerAPIVersion1)
		opt := options.Client().ApplyURI(mongoCfg.Source).SetServerAPIOptions(serverApi)
		c, err := mongo.Connect(context.TODO(), opt)

		if err != nil {
			panic(err)
		}

		client = c

		var result bson.M
		dbDashFun = client.Database(mongoCfg.DataBase)
		err = dbDashFun.RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result)

		if err != nil {
			panic(err)
		}
	}
	return dbDashFun
}

type IndexInfo struct {
	FieldName string
	Unique    bool
	Sort      int //升序=1，降序=-1
	IndexName string
}

func CreateIndexes(c *mongo.Collection, indexes []IndexInfo) error {
	cursor, err := c.Indexes().List(context.TODO())
	if err != nil {
		return err
	}
	defer cursor.Close(context.TODO())

	isExist := map[string]bool{}

	var index bson.M
	for cursor.Next(context.TODO()) {
		err = cursor.Decode(&index)
		if err != nil {
			return err
		}
		isExist[index["name"].(string)] = true
	}

	var indexModels []mongo.IndexModel

	for _, indexInfo := range indexes {
		if !isExist[indexInfo.IndexName] {
			indexModels = append(indexModels, mongo.IndexModel{
				Keys:    bson.D{{Key: indexInfo.FieldName, Value: indexInfo.Sort}},
				Options: options.Index().SetUnique(indexInfo.Unique).SetName(indexInfo.IndexName),
			})
		}
	}
	if len(indexModels) > 0 {
		_, err = c.Indexes().CreateMany(context.TODO(), indexModels)
		if err != nil {
			return err
		}
	}

	return nil
}
