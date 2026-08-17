package usercenter

import (
	"context"
	"dashfun_gamecenter/accountcenter"
	"dashfun_gamecenter/apperrors"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/snowflake"
	"dashfun_gamecenter/tencentcos"
	"dashfun_gamecenter/tgbot"
	"dashfun_gamecenter/utils"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/allegro/bigcache/v3"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"github.com/tonkeeper/tongo/ton"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// UserCenter TODO 定时清理超时用户，下线用户，并发送用户下线事件
type UserCenter struct {
	onlineUsers *OnlineUsers
	idGen       *snowflake.Worker
	avatarCache *bigcache.BigCache
}

func parseInitData(tgAuthData string, expIn time.Duration) (*initdata.InitData, error) {

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

func (uc *UserCenter) init() {
	uc.onlineUsers = newOnlineUsers()
	uc.idGen = snowflake.Must(snowflake.GetWorker(data.WorkerUserId))
	cfg := bigcache.DefaultConfig(1 * time.Hour)
	cfg.CleanWindow = 30 * time.Minute
	cache, err := bigcache.New(context.Background(), cfg)
	if err != nil {
		log.Panicln(err.Error())
	}
	uc.avatarCache = cache
}

func (uc *UserCenter) newUserId() string {
	id := uc.idGen.NextId()
	return "ur" + strconv.FormatInt(id, 36)
}

func (uc *UserCenter) RequestUserId() string {
	return uc.newUserId()
}

// restoreUserProfile keeps compatibility with profiles created before nickname
// and avatar fields were persisted directly on user_data.
func (uc *UserCenter) restoreUserProfile(user *data.DashFunUser) error {
	if user == nil || (user.Nickname != "" && user.AvatarUrl != "") {
		return nil
	}
	profile, err := dao.GetUserProfileDao().GetUserProfile(user.Id)
	if err != nil || profile == nil {
		return err
	}
	if user.Nickname == "" {
		user.Nickname = profile.Nickname
	}
	if user.AvatarUrl == "" {
		user.AvatarUrl = profile.Avatar
	}
	return nil
}

func (uc *UserCenter) GetDashFunUserChannelId(userId string, from data.DashFunUserFrom) (string, error) {
	user, err := uc.GetDashFunUser(userId)
	if err != nil {
		return "", err
	}
	return user.ChannelId, nil
}

// UserEnterGame 用户点击了Play按钮进入游戏
func (uc *UserCenter) UserEnterGame(authData *utils.AuthData, gameId string) (*data.DashFunUser, error) {
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

// UserLogin 用户登录, 通过tgAuthData获取用户信息, autoCreate表示是否自动创建用户
// 目前在只支持tma
func (uc *UserCenter) UserLogin(authData *utils.AuthData, referrerId string, autoCreate bool) (*data.OnlineUser, error) {
	if isAccountAuth(authData.Method) {
		return uc.accountUserLogin(authData.Token, referrerId, autoCreate)
	}
	if !isTelegramAuth(authData.Method) {
		return nil, errors.New("unsupported authorization method")
	}
	tgAuthData := authData.Token
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
	if err = uc.restoreUserProfile(user); err != nil {
		return nil, err
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

	return ou, nil
}

func isTelegramAuth(method string) bool {
	method = strings.ToLower(strings.TrimSpace(method))
	return method == "tma" || method == "telegram"
}

func isAccountAuth(method string) bool {
	method = strings.ToLower(strings.TrimSpace(method))
	return method == "bearer" || method == "account" || method == "duc"
}

func (uc *UserCenter) accountUserLogin(token, referrerId string, autoCreate bool) (*data.OnlineUser, error) {
	account, err := accountcenter.Get().AuthenticateToken(token)
	if err != nil {
		return nil, err
	}
	ud := dao.GetUserDao()
	user, err := ud.GetUserByChannelId(account.AccountId)
	if err != nil {
		return nil, err
	}
	newCreated := false
	if user == nil {
		if !autoCreate {
			return nil, apperrors.ErrUserDoesNotExist
		}
		now := time.Now().UnixMilli()
		user = &data.DashFunUser{Id: uc.newUserId(), ChannelId: account.AccountId, DisplayName: account.DisplayName, UserName: account.Username, From: data.DF_UserFrom_UserCenter, CreateData: now, LoginTime: now, WalletAddress: make(map[string]string)}
		if _, err = ud.SaveOrUpdate(user); err != nil {
			return nil, err
		}
		newCreated = true
	} else {
		user.LoginTime = time.Now().UnixMilli()
		user.DisplayName = account.DisplayName
		user.UserName = account.Username
	}
	if err = uc.restoreUserProfile(user); err != nil {
		return nil, err
	}

	var records []*data.PlayGameRecord
	var favorites []string
	if record, _ := dao.GetUserPlayRecordDao().GetUserPlayRecord(user.Id); record != nil {
		records, favorites = record.Records, record.Favorites
	}
	if records == nil {
		records = make([]*data.PlayGameRecord, 0)
	}
	if favorites == nil {
		favorites = make([]string, 0)
	}
	ou := uc.onlineUsers.TGUserLogin(user, &data.TGInfo{AuthData: ""}, records, favorites)

	if referrerId != "" && referrerId != user.Id && user.ReferrerId == "" {
		if referrer, refErr := uc.GetDashFunUser(referrerId); refErr == nil {
			user.ReferrerId = referrerId
			events.UserReferrerEvents.Emit(&events.UserReferrerEvent{User: user, Referrer: referrer, IsNewCreate: newCreated})
		}
	}
	if _, err = ud.SaveOrUpdate(user); err != nil {
		return nil, err
	}
	events.UserLoginEvents.Emit(ou)
	return ou, nil
}

func (uc *UserCenter) updateUserAvatar(user *data.DashFunUser) {
	photoFile := uc.getUserAvatarUrl(user)
	user.AvatarUrl = photoFile
}

func (uc *UserCenter) getUserAvatarUrl(user *data.DashFunUser) string {
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
func (uc *UserCenter) GetDashFunUserByAuthData(authData *utils.AuthData, onlineUserOnly bool) (*data.DashFunUser, error) {
	if isAccountAuth(authData.Method) {
		account, err := accountcenter.Get().AuthenticateToken(authData.Token)
		if err != nil {
			return nil, err
		}
		if ou := uc.onlineUsers.FindUserByChannelId(account.AccountId); ou != nil {
			return ou.User, nil
		}
		if onlineUserOnly {
			return nil, apperrors.ErrUserDoesNotExist
		}
		return dao.GetUserDao().GetUserByChannelId(account.AccountId)
	}
	if !isTelegramAuth(authData.Method) {
		return nil, errors.New("unsupported authorization method")
	}
	initData, err := parseInitData(authData.Token, 0)
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

func (uc *UserCenter) getUserPhoto(photoPath string) ([]byte, error) {
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

func (uc *UserCenter) GetUserChannelHeadData(userId string) []byte {
	avatarCached, err := uc.avatarCache.Get(userId)
	if err == nil {
		return avatarCached
	}
	user, err := uc.GetDashFunUser(userId)
	if err != nil {
		zap.S().Errorw("User Not Found", "userId", userId)
		return nil
	}
	photoPath := uc.getUserAvatarUrl(user)
	bytes, err := uc.getUserPhoto(photoPath)
	if errors.Is(err, apperrors.ErrUserPhotoNotExist) {
		return nil
	}
	uc.avatarCache.Set(userId, bytes)
	return bytes
}

func (uc *UserCenter) GetUserHeadAvatar(userId string) []byte {
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

func (uc *UserCenter) CreateDashFunUser(from data.DashFunUserFrom, username string) (*data.DashFunUser, error) {
	if from < data.DF_UserFrom_Kol {
		return nil, apperrors.ErrUserCannotCreateOnChannel
	}
	user := &data.DashFunUser{
		Id:          uc.newUserId(),
		ChannelId:   "",
		DisplayName: username,
		UserName:    username,
		AvatarUrl:   "",
		From:        from,
		CreateData:  time.Now().UnixMilli(),
		LoginTime:   time.Now().UnixMilli(),
		LogoffTime:  0,
	}
	user, err := dao.GetUserDao().SaveOrUpdate(user)
	return user, err
}

func (uc *UserCenter) GetUsersFrom(from data.DashFunUserFrom) ([]*data.DashFunUser, error) {
	if from < data.DF_UserFrom_Kol {
		return nil, apperrors.ErrUserCannotCreateOnChannel
	}
	users, err := dao.GetUserDao().GetUsersFrom(from)
	if err != nil {
		zap.S().Errorw("get users from error", "from", from, "err", err)
		return nil, err
	}
	return users, nil
}

func (uc *UserCenter) UserUpdateProfile(userId string, nickname string, avatar []byte) (*data.DashFunUser, error) {
	ou := uc.onlineUsers.FindUser(userId)
	if ou == nil {
		zap.S().Errorw("User Not Found", "userId", userId)
		return nil, apperrors.ErrOnlineUserNotExist
	}

	//只有用户没有设置昵称的时候才更新昵称
	if ou.User.Nickname == "" && nickname != "" {
		ou.User.Nickname = nickname
	}

	//avatar的存储方式是存储当前的版本号，文件路径固定是images/users/avatar_{userId}.png
	if len(avatar) > 0 {
		//upload avatar数据到tencent cos
		_, err := tencentcos.Get().UploadData("images/users/avatar_"+userId+".png", avatar, "image/png")
		if err != nil {
			return nil, err
		}

		if ou.User.AvatarUrl == "" {
			ou.User.AvatarUrl = "v1"
		} else {
			if strings.HasPrefix(ou.User.AvatarUrl, "v") {
				numStr := ou.User.AvatarUrl[1:]
				num, err := strconv.Atoi(numStr)
				if err != nil {
					ou.User.AvatarUrl = "v1"
				} else {
					ou.User.AvatarUrl = "v" + strconv.Itoa(num+1)
				}
			} else {
				ou.User.AvatarUrl = "v1"
			}
		}
	}

	if _, err := dao.GetUserDao().SaveOrUpdate(ou.User); err != nil {
		return nil, err
	}
	if _, err := dao.GetUserProfileDao().SaveOrUpdate(&data.UserProfileData{
		UserId:   userId,
		Nickname: ou.User.Nickname,
		Avatar:   ou.User.AvatarUrl,
	}); err != nil {
		return nil, err
	}

	return ou.User, nil
}
