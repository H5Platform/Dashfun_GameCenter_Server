package usercenter

import (
	"context"
	"dashfun_gamecenter/apperrors"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/nacoscenter"
	"dashfun_gamecenter/snowflake"
	"dashfun_gamecenter/tgbot"
	"dashfun_gamecenter/utils"
	"dashfun_gamecenter/utils/cache"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/allegro/bigcache/v3"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"github.com/tonkeeper/tongo/ton"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// UserCenterRpc TODO 定时清理超时用户，下线用户，并发送用户下线事件
type UserCenterRpc struct {
	onlineUsers *OnlineUsers
	idGen       *snowflake.Worker
	avatarCache *bigcache.BigCache

	// Auth信息缓存
	auth2user *cache.GenericCache[*data.DashFunUser]

	userCenterRpc *nacoscenter.UserCenterRpc
}

var onceUserCenterRpc sync.Once
var instUserCenterRpc *UserCenterRpc

func GetRpc() *UserCenterRpc {
	onceUserCenterRpc.Do(func() {
		instUserCenterRpc = &UserCenterRpc{}
		instUserCenterRpc.init()
	})
	return instUserCenterRpc
}

func parseInitDataRpc(tgAuthData string, expIn time.Duration) (*initdata.InitData, error) {

	token := config.GetConfig().TG.Token
	testIdx := strings.LastIndex(token, "/test")
	if testIdx >= 0 {
		token = token[:testIdx]
	}
	initData, err := initdata.Parse(tgAuthData)
	if err != nil {
		return nil, err
	}

	if !config.IsProd() {
		//非生产环境下，如果FirstName==Test并且id以999开头，则不验证，且做为测试账户登录
		if initData.User.FirstName == "Test" && strings.HasPrefix(strconv.Itoa(int(initData.User.ID)), "999") {
			return &initData, nil
		}
	}

	err = initdata.Validate(tgAuthData, token, expIn)
	if err != nil {
		return nil, err
	}

	return &initData, nil
}

func (uc *UserCenterRpc) init() {
	uc.onlineUsers = newOnlineUsers()
	uc.idGen = snowflake.Must(snowflake.GetWorker(data.WorkerUserId))
	cfg := bigcache.DefaultConfig(1 * time.Hour)
	cfg.CleanWindow = 30 * time.Minute
	avtCache, err := bigcache.New(context.Background(), cfg)
	if err != nil {
		log.Panicln(err.Error())
	}
	uc.avatarCache = avtCache
	uc.auth2user = cache.NewCache[*data.DashFunUser](5 * time.Minute)
	uc.userCenterRpc = nacoscenter.GetUserCenterRpc()
}

func (uc *UserCenterRpc) newUserId() string {
	id := uc.idGen.NextId()
	return "ur" + strconv.FormatInt(id, 36)
}

func (uc *UserCenterRpc) RequestUserId() string {
	return uc.newUserId()
}

func (uc *UserCenterRpc) GetDashFunUserChannelId(userId string, from data.DashFunUserFrom) (string, error) {
	return uc.userCenterRpc.GetUserChannelId(userId, from)
}

// UserEnterGame 用户点击了Play按钮进入游戏
func (uc *UserCenterRpc) UserEnterGame(authData *utils.AuthData, gameId string) (*data.DashFunUser, error) {
	user, err := uc.GetDashFunUserByAuthData(authData, false)

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

// TGUserLogin1 用户登录, 通过tgAuthData获取用户信息, autoCreate表示是否自动创建用户
// Deprecated: 废弃了
func (uc *UserCenterRpc) TGUserLogin1(tgAuthData string, referrerId string, autoCreate bool) (*data.OnlineUser, error) {
	initData, err := parseInitData(tgAuthData, 0)
	if err != nil {
		return nil, err
	}

	ud := dao.GetUserDao()
	user, err := ud.GetUserByChannelId(fmt.Sprintf("%d", initData.User.ID))
	if err != nil {
		return nil, err
	}

	if user == nil && !autoCreate {
		return nil, apperrors.ErrUserDoesNotExist
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
		uc.updateUserAvatar(user)
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

func (uc *UserCenterRpc) UserLogin(authData *utils.AuthData, referrerId string, autoCreate bool) (*data.OnlineUser, error) {
	user, newCreated, err := uc.userCenterRpc.UserLogin(authData, referrerId, autoCreate)
	if err != nil {
		return nil, err
	}

	var referrer *data.DashFunUser

	if referrerId != "" {
		referrer, err = uc.GetDashFunUser(referrerId)
		if err != nil || referrer == nil {
			referrerId = ""
		}
	} else {
		referrer = nil
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
		AuthData: authData.Token,
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
			IsNewCreate: newCreated,
		})
	}

	events.UserLoginEvents.Emit(ou)

	zap.S().Infow("User from Telegram Login Successful", "userId", user.Id, "tgUserId", user.ChannelId, "name", user.UserName)

	if !config.IsProd() {
		zap.S().Info("user tg token:", "auth", authData)
	}

	return ou, nil
}

func (uc *UserCenterRpc) updateUserAvatar(user *data.DashFunUser) {
	photoFile := uc.getUserAvatarUrl(user)
	user.AvatarUrl = photoFile
}

func (uc *UserCenterRpc) getUserAvatarUrl(user *data.DashFunUser) string {
	tgId, err := strconv.ParseInt(user.ChannelId, 10, 64)
	if err != nil {
		return ""
	}
	photoFile := tgbot.Get().GetUserPhotoFilePath(tgId)
	if photoFile != "" {
		photoFile = "TG-" + photoFile
	} else {
		photoFile = ""
	}
	return photoFile
}

// GetDashFunUserByAuthData 根据用户的tgAuthData，找到对应的DashFunUser
// onlineUserOnly -- 是否只检查在线用户
func (uc *UserCenterRpc) GetDashFunUserByAuthData(authData *utils.AuthData, onlineUserOnly bool) (*data.DashFunUser, error) {
	user, hit := uc.auth2user.Get(authData.ToString())
	if hit {
		return user, nil
	}
	user, err := uc.userCenterRpc.GetDashFunUserByAuthData(authData, onlineUserOnly)
	if err != nil {
		return nil, err
	}
	uc.auth2user.Set(authData.ToString(), user)
	return user, nil
}

func (uc *UserCenterRpc) GetDashFunUser(userId string) (*data.DashFunUser, error) {
	ou := uc.onlineUsers.FindUser(userId)
	var user *data.DashFunUser
	if ou == nil {
		uf, err := uc.userCenterRpc.GetDashFunUser(userId)
		if err != nil && !errors.Is(err, apperrors.ErrUserNotFound) {
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

func (uc *UserCenterRpc) UserBindWallet(userId, chain, address string) (*data.DashFunUser, error) {
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

func (uc *UserCenterRpc) getDataKey(dataKey string, isTesting bool) string {
	key := dataKey
	if isTesting {
		key = "TEST_" + key
	}
	return key
}

// UserSaveData 保存用户数据
func (uc *UserCenterRpc) UserSaveData(userId, gameId, dataKey, saveData string, isTesting bool) (*data.DashFunUserSaveData, error) {
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
func (uc *UserCenterRpc) UserGetData(userId, gameId, dataKey string, isTesting bool) (string, error) {
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
func (uc *UserCenterRpc) UserGetFavorites(userId string) []string {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		zap.S().Errorw("User Not Found", "userId", userId)
		return nil
	}
	return ou.Favorites
}

// UserAddFavorite adds a game to the user's favorites
func (uc *UserCenterRpc) UserAddFavorite(userId, gameId string) error {
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
func (uc *UserCenterRpc) IsUserFavoriteGame(userId, gameId string) (bool, error) {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		return false, apperrors.ErrOnlineUserNotExist
	}
	return ou.IsFavoriteGame(gameId), nil
}

// UserRemoveFavorite removes a game from the user's favorites
func (uc *UserCenterRpc) UserRemoveFavorite(userId, gameId string) error {
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

func (uc *UserCenterRpc) UserGetPlayRecord(userId string) []*data.PlayGameRecord {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		zap.S().Errorw("User Not Found", "userId", userId)
		return nil
	}
	return ou.PlayRecord
}

func (uc *UserCenterRpc) getUserPhoto(photoPath string) ([]byte, error) {
	if photoPath == "" {
		return nil, apperrors.ErrUserPhotoNotExist
	}
	prefix := photoPath[:3]
	filePath := photoPath[3:]
	if prefix == "TG-" {
		filePath = tgbot.Get().GetUserPhotoUrlByFile(filePath)
		zap.S().Infow("Get User Photo", "filePath", filePath)
		resp, err := http.Get(filePath)
		if err == nil {
			defer resp.Body.Close()
			d, err := io.ReadAll(resp.Body)
			if err == nil {
				str := string(d)
				if strings.HasPrefix(str, "{\"ok\":false") {
					zap.S().Errorw("User Photo Not Found", "filePath", filePath, "data", str)
					//用户头像不存在，需要更新
					return nil, apperrors.ErrUserPhotoNotExist
				}
				return d, nil
			}
		}
	}
	return nil, nil
}

func (uc *UserCenterRpc) GetUserChannelHeadData(userId string) []byte {
	avatarCached, err := uc.avatarCache.Get(userId)
	if err == nil {
		return avatarCached
	}
	avatar, err := uc.userCenterRpc.GetUserAvatar(userId)
	if err != nil {
		return nil
	}
	return avatar
}

func (uc *UserCenterRpc) GetUserHeadAvatar(userId string) []byte {
	avatarCached, err := uc.avatarCache.Get(userId)
	if err == nil {
		return avatarCached
	}
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		zap.S().Errorw("User Not Found", "userId", userId)
		return nil
	}

	var avatar []byte = nil

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
				uc.avatarCache.Set(userId, d)
				avatar = d
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

	//return ou.Header
	return avatar
}
