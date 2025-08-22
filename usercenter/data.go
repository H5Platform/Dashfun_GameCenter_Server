package usercenter

import (
	"context"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/datasource/mongoimpl"
	"dashfun_gamecenter/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"sync"
	"time"
)

type OnlineUsers struct {
	// key = userId
	Users map[string]*data.OnlineUser
	//channel_id -> userId
	ChannelMap map[string]string
	sync.RWMutex
}

func newOnlineUsers() *OnlineUsers {
	return &OnlineUsers{
		Users:      make(map[string]*data.OnlineUser),
		ChannelMap: make(map[string]string),
	}
}

func (o *OnlineUsers) FindUserByChannelId(channelId string) *data.OnlineUser {
	o.RLock()
	defer o.RUnlock()

	userId, exist := o.ChannelMap[channelId]
	if !exist {
		return nil
	}
	user, exist := o.Users[userId]
	if !exist {
		return nil
	}
	return user
}

func (o *OnlineUsers) FindUser(userId string) *data.OnlineUser {
	o.RLock()
	defer o.RUnlock()

	user, exist := o.Users[userId]
	if !exist {
		return nil
	}
	return user
}

func (o *OnlineUsers) TGUserLogin(user *data.DashFunUser, tgInfo *data.TGInfo, playRecord []*data.PlayGameRecord, favorites []string) *data.OnlineUser {
	o.Lock()
	defer o.Unlock()

	u, e := o.Users[user.Id]
	if !e {
		u = data.NewOnlineUser(user, tgInfo, playRecord, favorites)
		o.Users[user.Id] = u
	}
	o.ChannelMap[user.ChannelId] = user.Id
	u.User.LoginTime = time.Now().UnixMilli()
	u.Header = nil
	return u
}

func (o *OnlineUsers) UserLogout(user *data.DashFunUser) *data.OnlineUser {
	o.Lock()
	defer o.Unlock()

	u, e := o.Users[user.Id]
	if e {
		u.User.LogoffTime = time.Now().UnixMilli()
		delete(o.Users, user.Id)
		delete(o.ChannelMap, user.ChannelId)
	}
	return u
}

var onceUserCenter sync.Once
var instUserCenter IUserCenter

type IUserCenter interface {
	UserEnterGame(authData *utils.AuthData, gameId string) (*data.DashFunUser, error)
	UserLogin(authData *utils.AuthData, referrerId string, autoCreate bool) (*data.OnlineUser, error)
	GetDashFunUserByAuthData(authData *utils.AuthData, onlineUserOnly bool) (*data.DashFunUser, error)
	GetDashFunUserChannelId(userId string, from data.DashFunUserFrom) (string, error)
	GetDashFunUser(userId string) (*data.DashFunUser, error)
	UserBindWallet(userId, chain, address string) (*data.DashFunUser, error)
	UserSaveData(userId, gameId, dataKey, saveData string, isTesting bool) (*data.DashFunUserSaveData, error)
	UserGetData(userId, gameId, dataKey string, isTesting bool) (string, error)
	UserGetFavorites(userId string) []string
	UserAddFavorite(userId, gameId string) error
	IsUserFavoriteGame(userId, gameId string) (bool, error)
	UserRemoveFavorite(userId, gameId string) error
	UserGetPlayRecord(userId string) []*data.PlayGameRecord
	GetUserChannelHeadData(userId string) []byte
	GetUserHeadAvatar(userId string) []byte
	RequestUserId() string
	CreateDashFunUser(from data.DashFunUserFrom, username string) (*data.DashFunUser, error)
	GetUsersFrom(from data.DashFunUserFrom) ([]*data.DashFunUser, error)
	UserUpdateProfile(userId string, nickname string, avatar []byte) (*data.DashFunUser, error)
}

func Get() IUserCenter {
	onceUserCenter.Do(func() {
		uc := &UserCenterRpc{}
		uc.init()
		instUserCenter = uc
	})
	return instUserCenter
}

// MoveUserData 将用户数据迁移到新的数据源，新的数据源供单独的UserCenter服务使用
func MoveUserData() {

	mongoCfg := config.GetConfig().Mongo
	serverApi := options.ServerAPI(options.ServerAPIVersion1)
	opt := options.Client().ApplyURI(mongoCfg.Source).SetServerAPIOptions(serverApi)
	client, err := mongo.Connect(context.TODO(), opt)

	if err != nil {
		panic(err)
	}

	var result bson.M
	dbDashFun := client.Database(mongoCfg.DataBase)
	err = dbDashFun.RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result)

	if err != nil {
		panic(err)
	}

	dbUserCenter := client.Database("DBUserCenter")
	err = dbUserCenter.RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result)

	if err != nil {
		panic(err)
	}

	//检查数据库是否有user_data集合，如果有则说明已经迁移过了
	exists, err := mongoimpl.CollectionExists(context.Background(), dbUserCenter, "user_data")
	if err != nil {
		panic(err)
	}

	if exists {
		return
	}

	log.Printf("start moving user data")

	cursor, err := dbDashFun.Collection("user_data").Find(context.TODO(), bson.M{})

	if err != nil {
		panic(err)
	}

	users := make([]*data.DashFunUser, 0)
	err = cursor.All(context.TODO(), &users)
	if err != nil {
		panic(err)
	}

	cursor.Close(context.Background())

	log.Printf("user data moving users %d", len(users))

	for _, user := range users {
		update := bson.M{
			"$set": user,
		}
		opts := options.Update().SetUpsert(true)
		_, err := dbUserCenter.Collection("user_data").UpdateByID(context.TODO(), user.Id, update, opts)
		if err != nil {
			log.Printf("insert user data error %s", err)
		}

		update = bson.M{
			"$set": &bson.D{
				bson.E{Key: "_id", Value: user.Id},
				bson.E{Key: "channel_id", Value: user.ChannelId},
				bson.E{Key: "from", Value: data.DF_UserFrom_TG},
				bson.E{Key: "auth_data", Value: ""},
			},
		}

		_, err = dbUserCenter.Collection("user_channel_data").UpdateByID(context.TODO(), user.Id, update, opts)
		if err != nil {
			log.Panicf("insert user data error %s\n", err)
		}
	}

}
