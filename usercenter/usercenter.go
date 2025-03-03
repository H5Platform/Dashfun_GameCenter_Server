package usercenter

import (
	"dashfun_gamecenter/apperrors"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/snowflake"
	"dashfun_gamecenter/tgbot"
	"encoding/base64"
	"errors"
	"fmt"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"github.com/tonkeeper/tongo/ton"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"io"
	"net/http"
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

func (uc *UserCenter) RequestUserId() string {
	return uc.newUserId()
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

	ou := uc.onlineUsers.FindUser(user.Id)
	ou.AddPlayRecord(gameId)

	dao.GetUserPlayRecordDao().SaveOrUpdate(&data.DashFunUserPlayRecord{
		UserId:    user.Id,
		Records:   ou.PlayRecord,
		Favorites: ou.Favorites,
	})

	events.UserEnterGameEvents.Emit(&events.EventUserEnterGame{
		User: user,
		Game: game,
	})

	zap.S().Infow("User from Telegram Enter Game", "userId", user.Id, "tgUserId", user.ChannelId, "name", user.UserName, "game", gameId)

	return user, nil
}

func (uc *UserCenter) TGUserLogin(tgAuthData string, referrerId string) (*data.OnlineUser, error) {
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
	//photoUrl := tgbot.Get().GetUserPhotoUrl(tgUser.ID)

	var referrer *data.DashFunUser

	if referrerId != "" {
		referrer, err = uc.GetDashFunUser(referrerId)
		if err != nil || referrer == nil {
			referrerId = ""
		}
	} else {
		referrer = nil
	}

	newCreate := false

	if user == nil {
		//create new user
		user = &data.DashFunUser{
			Id:          uc.newUserId(),
			ChannelId:   fmt.Sprintf("%d", tgUser.ID),
			DisplayName: fmt.Sprintf("%s %s", tgUser.FirstName, tgUser.LastName),
			UserName:    tgUser.Username,
			AvatarUrl:   "",
			From:        data.DF_UserFrom_TG,
			CreateData:  time.Now().UnixMilli(),
			LoginTime:   time.Now().UnixMilli(),
			LogoffTime:  0,
		}
		_, err = ud.SaveOrUpdate(user)
		newCreate = true
		if err != nil {
			zap.S().Errorw("save user error", "user", user, "err", err)
			return nil, err
		} else {
			zap.S().Debugw("User Created", "user", user)
		}
	} else {
		user.LoginTime = time.Now().UnixMilli()
		//update display name
		user.DisplayName = fmt.Sprintf("%s %s", tgUser.FirstName, tgUser.LastName)
		//update avatar info
		photoFile := tgbot.Get().GetUserPhotoFilePath(tgUser.ID)
		if photoFile != "" {
			user.AvatarUrl = "TG-" + photoFile
		} else {
			user.AvatarUrl = ""
		}
	}

	var playRecord []*data.PlayGameRecord
	var favorites []string

	record, err := dao.GetUserPlayRecordDao().GetUserPlayRecord(user.Id)
	if record == nil || record.Records == nil {
		playRecord = make([]*data.PlayGameRecord, 0)
	} else {
		playRecord = record.Records
	}

	if record == nil || record.Favorites == nil {
		favorites = make([]string, 0)
	} else {
		favorites = record.Favorites
	}

	ou := uc.onlineUsers.TGUserLogin(user, &data.TGInfo{
		AuthData: tgAuthData,
		InitData: initData,
	}, playRecord, favorites)

	if referrerId != "" && user.Id != referrerId {
		//check referrer
		if user.ReferrerId == "" {
			//用户没有推荐人，设置推荐人 (由于用户可被多次邀请，此处只是记录首次邀请人id)
			user.ReferrerId = referrerId
		}
		events.UserReferrerEvents.Emit(&events.UserReferrerEvent{
			User:        user,
			Referrer:    referrer,
			IsNewCreate: newCreate,
		})
	}
	_, err = ud.SaveOrUpdate(user)
	if err != nil {
		zap.S().Errorw("save user error", "user", user, "err", err)
		return nil, err
	}

	events.UserLoginEvents.Emit(ou)

	zap.S().Infow("User from Telegram Login Successful", "userId", user.Id, "tgUserId", user.ChannelId, "name", user.UserName)

	if !config.IsProd() {
		zap.S().Info("user tg token:", tgAuthData)
	}

	return ou, nil
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
		return nil, apperrors.ErrUserDoesNotExist
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

func (uc *UserCenter) getDataKey(dataKey string, isTesting bool) string {
	key := dataKey
	if isTesting {
		key = "TEST_" + key
	}
	return key
}

// UserSaveData 保存用户数据
func (uc *UserCenter) UserSaveData(userId, gameId, dataKey, saveData string, isTesting bool) (*data.DashFunUserSaveData, error) {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		//只有在线用户给保存数据
		return nil, apperrors.ErrOnlineUserNotExist
	}

	key := uc.getDataKey(dataKey, isTesting)

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
func (uc *UserCenter) UserGetData(userId, gameId, dataKey string, isTesting bool) (string, error) {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		//只有在线用户给读取数据
		return "", apperrors.ErrOnlineUserNotExist
	}

	key := uc.getDataKey(dataKey, isTesting)

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

// UserGetFavorites 获取用户收藏的游戏
func (uc *UserCenter) UserGetFavorites(userId string) []string {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		zap.S().Errorw("User Not Found", "userId", userId)
		return nil
	}
	return ou.Favorites
}

// UserAddFavorite adds a game to the user's favorites
func (uc *UserCenter) UserAddFavorite(userId, gameId string) error {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		return apperrors.ErrOnlineUserNotExist
	}
	ou.AddFavoriteGame(gameId)
	// Save the updated favorites to the database
	_, err := dao.GetUserPlayRecordDao().SaveOrUpdate(&data.DashFunUserPlayRecord{
		UserId:    userId,
		Records:   ou.PlayRecord,
		Favorites: ou.Favorites,
	})
	if err != nil {
		return err
	}

	return nil
}

// IsUserFavoriteGame checks if a game is in the user's favorites
func (uc *UserCenter) IsUserFavoriteGame(userId, gameId string) (bool, error) {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		return false, apperrors.ErrOnlineUserNotExist
	}
	return ou.IsFavoriteGame(gameId), nil
}

// UserRemoveFavorite removes a game from the user's favorites
func (uc *UserCenter) UserRemoveFavorite(userId, gameId string) error {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		return apperrors.ErrOnlineUserNotExist
	}
	ou.RemoveFavoriteGame(gameId)
	// Save the updated favorites to the database
	_, err := dao.GetUserPlayRecordDao().SaveOrUpdate(&data.DashFunUserPlayRecord{
		UserId:    userId,
		Records:   ou.PlayRecord,
		Favorites: ou.Favorites,
	})
	if err != nil {
		return err
	}

	return nil
}

func (uc *UserCenter) UserGetPlayRecord(userId string) []*data.PlayGameRecord {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		zap.S().Errorw("User Not Found", "userId", userId)
		return nil
	}
	return ou.PlayRecord
}

func (uc *UserCenter) GetUserChannelHeadData(userId, photoPath string) []byte {
	prefix := photoPath[:3]
	filePath := photoPath[3:]
	if prefix == "TG-" {
		filePath = tgbot.Get().GetUserPhotoUrlByFile(filePath)
		resp, err := http.Get(filePath)
		if err == nil {
			defer resp.Body.Close()
			d, err := io.ReadAll(resp.Body)
			if err == nil {
				return d
			}
		}
	}
	return nil
}

func (uc *UserCenter) GetUserHeadAvatar(userId string) []byte {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		zap.S().Errorw("User Not Found", "userId", userId)
		return nil
	}
	//暂时不做缓存
	// if ou.Header == nil {
	// photoUrl := tgbot.Get().GetUserPhotoUrl(ou.TGInfo.InitData.User.ID)

	photoUrl := ou.User.AvatarUrl
	if photoUrl == "" {
		//用户没有头像
	} else {
		//prefix还没用，后面可能会根据prefix区分渠道，使用不同的头像获取方式
		//prefix := photoUrl[:3]
		filePath := tgbot.Get().GetUserPhotoUrlByFile(photoUrl[3:])

		resp, err := http.Get(filePath)
		if err == nil {
			defer resp.Body.Close()
			d, err := io.ReadAll(resp.Body)
			if err == nil {
				ou.Header = d
			}
		}
	}
	//}

	//if ou.Header == nil {
	//	//根据用户id生成一个头像
	//	hash16 := md5.Sum([]byte(userId))
	//	hash := hash16[:]
	//	salt16 := md5.Sum([]byte(userId + "SALT"))
	//	salt := salt16[:]
	//
	//	icon := identicon.New7x7(salt).Render(hash)
	//	ou.Header = icon
	//}

	return ou.Header
}
