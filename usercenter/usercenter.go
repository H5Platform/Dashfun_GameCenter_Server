package usercenter

import (
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/snowflake"
	"errors"
	"fmt"
	initdata "github.com/telegram-mini-apps/init-data-golang"
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
	user, err := uc.GetDashFunUserByTgAuthData(tgAuthData)
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
	initData, err := parseInitData(tgAuthData, time.Hour)
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
	return onlineUser, nil
}

// GetDashFunUserByTgAuthData 根据用户的tgAuthData，找到对应的DashFunUser
func (uc *UserCenter) GetDashFunUserByTgAuthData(tgAuthData string) (*data.DashFunUser, error) {
	initData, err := parseInitData(tgAuthData, 0)
	if err != nil {
		return nil, err
	}

	channelId := strconv.FormatInt(initData.User.ID, 10)

	ou := uc.onlineUsers.FindUserByChannelId(channelId)
	var user *data.DashFunUser

	if ou == nil {
		uf, err := dao.GetUserDao().GetUserByChannelId(channelId)
		if err != nil {
			return nil, err
		}
		user = uf
	} else {
		user = ou.User
	}
	if user == nil {
		zap.S().Errorw("User Not Found By TGAuthData", "tgUser", initData.User)
		return nil, errors.New("user does not exist")
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
