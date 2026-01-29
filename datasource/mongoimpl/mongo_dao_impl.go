package mongoimpl

import (
	"context"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/types"
	"time"

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
	spinWheelDao          types.SpinWheelDao
	spinWheelUserDao      types.SpinWheelUserDao
	userSaveDataDao       types.DashFunUserSaveDataDao
	userPlayRecordDao     types.DashFunUserPlayRecordDao
	invitedUserDao        types.InvitedUserDao
	rechargeDao           types.RechargeDao
	leaderboardBotDao     types.LeaderboardBotDao
	pricePredictDao       types.PricePredictDao
	exchangeDao           types.ExchangeDao
	squadGameDao          types.SquadGameDao
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
		spinWheelDao:          GetSpinWheelDaoMongo(),
		spinWheelUserDao:      GetSpinWheelUserDaoMongo(),
		userSaveDataDao:       GetUserSaveDataDaoMongo(),
		userPlayRecordDao:     GetUserPlayRecordDaoMongo(),
		invitedUserDao:        GetInvitedUserDaoMongo(),
		rechargeDao:           GetRechargeDaoMongo(),
		leaderboardBotDao:     GetLeaderboardBotDaoMongo(),
		pricePredictDao:       GetPricePredictDaoMongo(),
		exchangeDao:           NewExchangeDaoMongo(GetMongoDatabase()),
		squadGameDao:          GetSquadGameDaoMongo(),
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
func (d *DaoImplMongo) GetSpinWheelDao() types.SpinWheelDao              { return d.spinWheelDao }
func (d *DaoImplMongo) GetSpinWheelUserDao() types.SpinWheelUserDao      { return d.spinWheelUserDao }
func (d *DaoImplMongo) GetUserSaveDataDao() types.DashFunUserSaveDataDao { return d.userSaveDataDao }
func (d *DaoImplMongo) GetUserPlayRecordDao() types.DashFunUserPlayRecordDao {
	return d.userPlayRecordDao
}
func (d *DaoImplMongo) GetInvitedUserDao() types.InvitedUserDao {
	return d.invitedUserDao
}
func (d *DaoImplMongo) GetRechargeDao() types.RechargeDao {
	return d.rechargeDao
}
func (d *DaoImplMongo) GetLeaderboardBotDao() types.LeaderboardBotDao {
	return d.leaderboardBotDao
}
func (d *DaoImplMongo) GetAirdropDao() types.AirdropDao {
	return GetAirdropDaoMongo()
}
func (d *DaoImplMongo) GetUserProfileDao() types.UserProfileDao {
	return GetUserProfileDaoMongo()
}
func (d *DaoImplMongo) GetFishingPostDao() types.FishingPostDao {
	return GetFishingPostDaoMongo()
}
func (d *DaoImplMongo) GetFishingLeaderboardBotDao() types.FishingLeaderboardBotDao {
	return GetFishingLeaderboardBotDaoMongo()
}
func (d *DaoImplMongo) GetNolanDevPostDao() types.NolanDevPostDao {
	return GetNolanDevPostDaoMongo()
}
func (d *DaoImplMongo) GetNolanDevLeaderboardBotDao() types.NolanDevLeaderboardBotDao {
	return GetNolanDevLeaderboardBotDaoMongo()
}

func (d *DaoImplMongo) GetPricePredictDao() types.PricePredictDao {
	return d.pricePredictDao
}

func (d *DaoImplMongo) GetExchangeDao() types.ExchangeDao {
	return d.exchangeDao
}

func (d *DaoImplMongo) GetSquadGameDao() types.SquadGameDao {
	return d.squadGameDao
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

type MongoCursor[T any] struct {
	c      *mongo.Collection
	cursor *mongo.Cursor
}

func (c *MongoCursor[T]) Next(ctx context.Context) bool {
	return c.cursor.Next(ctx)
}

func (c *MongoCursor[T]) Data() (*T, error) {
	var ret T
	err := c.cursor.Decode(&ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *MongoCursor[T]) Close() {
	c.cursor.Close(context.Background())
}

func newMongoCursor[T any](c *mongo.Collection, filter *bson.D, batchSize int32) (*MongoCursor[T], error) {
	mc := &MongoCursor[T]{
		c: c,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if filter == nil {
		filter = &bson.D{}
	}
	opt := options.Find().SetBatchSize(batchSize)
	cursor, err := c.Find(ctx, filter, opt)

	if err != nil {
		return nil, err
	}
	mc.cursor = cursor
	return mc, nil
}

func CollectionExists(ctx context.Context, db *mongo.Database, collectionName string) (bool, error) {
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return false, err
	}

	for _, name := range collections {
		if name == collectionName {
			return true, nil
		}
	}
	return false, nil
}

func DatabaseExists(ctx context.Context, client *mongo.Client, dbName string) (bool, error) {
	databases, err := client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return false, err
	}

	for _, name := range databases {
		if name == dbName {
			return true, nil
		}
	}
	return false, nil
}
