package usercenter

import (
	"dashfun_gamecenter/apperrors"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/snowflake"
	"encoding/base64"
	"errors"
	"fmt"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"github.com/tonkeeper/tongo/ton"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"time"
)

// UserCenter TODO 定时清理超时用户，下线用户，并发送用户下线事件
type UserCenter struct {
	onlineUsers *OnlineUsers
	idGen       *snowflake.Worker
}

var onceUserCenter sync.Once
var instUserCenter *UserCenter

func Get() *UserCenter {
	onceUserCenter.Do(func() {
		instUserCenter = &UserCenter{
			onlineUsers: newOnlineUsers(),
			idGen:       snowflake.Must(snowflake.GetWorker(data.WorkerUserId)),
		}
	})
	return instUserCenter
}

func parseInitData(tgAuthData string, expIn time.Duration) (*initdata.InitData, error) {
	token := config.GetConfig().TG.Token
	testIdx := strings.LastIndex(token, "/test")
	if testIdx >= 0 {
		token = token[:testIdx]
	}
	err := initdata.Validate(tgAuthData, token, expIn)
	if err != nil {
		return nil, err
	}

	initData, err := initdata.Parse(tgAuthData)
	if err != nil {
		return nil, err
	}

	return &initData, nil
}

func (uc *UserCenter) newUserId() string {
	id := uc.idGen.NextId()
	return strconv.FormatInt(id, 36)
}

// UserEnterGame 用户点击了Play按钮进入游戏
func (uc *UserCenter) UserEnterGame(tgAuthData, gameId string) (*data.DashFunUser, error) {
	user, err := uc.GetDashFunUserByTgAuthData(tgAuthData, false)
	if err != nil {
		return nil, err
	}
	game, err := gamecenter.Get().FindGame(gameId)
	if err != nil {
		return nil, err
	}
	events.UserEnterGameEvents.Emit(&events.EventUserEnterGame{
		User: user,
		Game: game,
	})

	zap.S().Infow("User from Telegram Enter Game", "userId", user.Id, "tgUserId", user.ChannelId, "name", user.UserName, "game", gameId)

	return user, nil
}

func (uc *UserCenter) TGUserLogin(tgAuthData string) (*data.OnlineUser, error) {
	initData, err := parseInitData(tgAuthData, 0)
	if err != nil {
		return nil, err
	}

	ud := dao.GetUserDao()
	user, err := ud.GetUserByChannelId(fmt.Sprintf("%d", initData.User.ID))
	if err != nil {
		return nil, err
	}

	tgUser := initData.User

	if user == nil {
		//create new user
		user = &data.DashFunUser{
			Id:          uc.newUserId(),
			ChannelId:   fmt.Sprintf("%d", tgUser.ID),
			DisplayName: fmt.Sprintf("%s %s", tgUser.FirstName, tgUser.LastName),
			UserName:    tgUser.Username,
			AvatarUrl:   tgUser.PhotoURL,
			From:        data.DF_UserFrom_TG,
			CreateData:  time.Now().UnixMilli(),
			LoginTime:   time.Now().UnixMilli(),
			LogoffTime:  0,
		}
		zap.S().Debugw("User Created", "user", user)
	}

	u, err := ud.SaveOrUpdate(user)
	if err != nil {
		zap.S().Errorw("save user error", "user", user, "err", err)
		return nil, err
	}

	onlineUser := uc.onlineUsers.TGUserLogin(u, &data.TGInfo{
		AuthData: tgAuthData,
		InitData: initData,
	})

	events.UserLoginEvents.Emit(onlineUser)

	zap.S().Infow("User from Telegram Login Successful", "userId", user.Id, "tgUserId", user.ChannelId, "name", user.UserName)

	if !config.IsProd() {
		zap.S().Info("user tg token:", tgAuthData)
	}

	return onlineUser, nil
}

// GetDashFunUserByTgAuthData 根据用户的tgAuthData，找到对应的DashFunUser
// onlineUserOnly -- 是否只检查在线用户
func (uc *UserCenter) GetDashFunUserByTgAuthData(tgAuthData string, onlineUserOnly bool) (*data.DashFunUser, error) {
	initData, err := parseInitData(tgAuthData, 0)
	if err != nil {
		return nil, err
	}

	channelId := strconv.FormatInt(initData.User.ID, 10)

	ou := uc.onlineUsers.FindUserByChannelId(channelId)
	var user *data.DashFunUser

	if ou == nil {
		//只检查在线用户的情况下不读取数据库
		if onlineUserOnly {
			user = nil
		} else {
			uf, err := dao.GetUserDao().GetUserByChannelId(channelId)
			if err != nil {
				return nil, err
			}
			user = uf
		}
	} else {
		user = ou.User
	}
	if user == nil {
		zap.S().Errorw("User Not Found By TGAuthData", "tgUser", initData.User)
		return nil, apperrors.ErrUserDoesNotExist
	}

	return user, nil
}

func (uc *UserCenter) GetDashFunUser(userId string) (*data.DashFunUser, error) {
	ou := uc.onlineUsers.FindUser(userId)
	var user *data.DashFunUser
	if ou == nil {
		uf, err := dao.GetUserDao().GetUserById(userId)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		user = uf
	} else {
		user = ou.User
	}
	if user == nil {
		zap.S().Errorw("User Not Found By UserId", "userId", userId)
		return nil, errors.New("user does not exist")
	}

	return user, nil
}

func (uc *UserCenter) UserBindWallet(userId, chain, address string) (*data.DashFunUser, error) {
	user, err := uc.GetDashFunUser(userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user does not exist")
	}

	if chain == "Ton" {
		acc, err := ton.ParseAccountID(address)
		zap.S().Infow("user bind wallet", "chain", chain, "address", address, "acc", acc.ToHuman(false, false))
		if err != nil {
			return nil, err
		}
	}

	if user.WalletAddress == nil {
		user.WalletAddress = make(map[string]string)
	}

	user.WalletAddress[chain] = address
	dao.GetUserDao().SaveOrUpdate(user)
	events.UserBindAddressEvents.Emit(&events.EventUserBindWallet{
		User:    user,
		Chain:   chain,
		Address: address,
	})
	return user, nil
}

// UserSaveData 保存用户数据
func (uc *UserCenter) UserSaveData(userId, gameId, key, saveData string) (*data.DashFunUserSaveData, error) {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		//只有在线用户给保存数据
		return nil, apperrors.ErrOnlineUserNotExist
	}
	dd := base64.StdEncoding.EncodeToString([]byte(saveData))
	ou.SetGameSaveData(gameId, key, dd)
	//同时存库
	d := dao.GetUserSaveDataDao()
	ret := &data.DashFunUserSaveData{
		UserId: userId,
		GameId: gameId,
		Key:    key,
		Data:   dd,
	}
	d.SaveOrUpdate(ret)
	return ret, nil
}

// UserGetData 用户获取保存的数据
// userId	用户Id
// gameId	游戏Id
// key		数据键值
func (uc *UserCenter) UserGetData(userId, gameId, key string) (string, error) {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		//只有在线用户给读取数据
		return "", apperrors.ErrOnlineUserNotExist
	}

	gameSaveData, err := ou.GetGameSaveData(gameId, key)
	if errors.Is(err, apperrors.ErrUserGameSaveDataNotExisted) {
		//用户数据不存在，尝试读取数据库
		d := dao.GetUserSaveDataDao()
		saveData, err := d.GetUserSaveData(userId, gameId, key)
		if err != nil {
			return "", err
		}
		if saveData == nil {
			//用户没有保存过数据，临时存储一条
			ou.SetGameSaveData(gameId, key, "")
		} else {
			ou.SetGameSaveData(gameId, key, saveData.Data)
			gameSaveData = saveData.Data
		}
	} else if err != nil {
		return "", err
	}

	return gameSaveData, nil
}
