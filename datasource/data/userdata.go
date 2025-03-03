package data

import (
	"dashfun_gamecenter/apperrors"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"slices"
	"time"
)

type DashFunUserFrom int

const (
	DF_UserFrom_TG DashFunUserFrom = iota + 1 //telegram 用户
)

// DashFunUser user data db
type DashFunUser struct {
	Id            string            `json:"id" bson:"_id"`                        //全局userId
	ChannelId     string            `json:"channel_id" bson:"channel_id"`         //渠道方id
	DisplayName   string            `json:"display_name" bson:"display_name"`     //显示名称
	UserName      string            `json:"user_name" bson:"user_name"`           //用户名
	AvatarUrl     string            `json:"avatar_url" bson:"avatar_url"`         //avatar地址
	From          DashFunUserFrom   `json:"from" bson:"from"`                     //用户来源
	CreateData    int64             `json:"create_data" bson:"create_data"`       //创建时间
	LoginTime     int64             `json:"login_time" bson:"login_time"`         //登录时间
	LogoffTime    int64             `json:"logoff_time" bson:"logoff_time"`       //登出时间
	WalletAddress map[string]string `json:"wallet_address" bson:"wallet_address"` //钱包地址 key=网络，value=地址
	ReferrerId    string            `json:"referrer_id" bson:"referrer_id"`       //推荐人id
}

type TGInfo struct {
	AuthData string
	InitData *initdata.InitData
}

type UserSaveData struct {
	GameId string
	//key->data
	SaveData map[string]string
}

func (usd *UserSaveData) GetSaveData(key string) string {
	v, ok := usd.SaveData[key]
	if ok {
		return v
	}
	return ""
}

func NewOnlineUser(User *DashFunUser, TGInfo *TGInfo, playRecord []*PlayGameRecord, favorites []string) *OnlineUser {
	return &OnlineUser{
		User:       User,
		TGInfo:     TGInfo,
		SaveData:   map[string]*UserSaveData{},
		PlayRecord: playRecord,
		Favorites:  favorites,
	}
}

type OnlineUser struct {
	User   *DashFunUser
	TGInfo *TGInfo
	//用户在各个游戏中存储的数据
	//GameId -> *UserSaveData
	SaveData map[string]*UserSaveData
	//用户的游戏记录
	PlayRecord []*PlayGameRecord
	//用户的收藏游戏id
	Favorites []string
	//用户头像缓存数据
	Header []byte
}

func (ou *OnlineUser) AddPlayRecord(gameId string) {
	found := false
	for _, record := range ou.PlayRecord {
		if record.GameId == gameId {
			found = true
			record.PlayTime = time.Now().UnixMilli()
			break
		}
	}
	if !found {
		ou.PlayRecord = append(ou.PlayRecord, &PlayGameRecord{
			GameId:   gameId,
			PlayTime: time.Now().UnixMilli(),
		})
	}

	slices.SortFunc(ou.PlayRecord, func(a, b *PlayGameRecord) int {
		return int(b.PlayTime - a.PlayTime)
	})

	if len(ou.PlayRecord) > 30 {
		ou.PlayRecord = ou.PlayRecord[0:30]
	}
}

// IsFavoriteGame checks if a game is in the user's favorites
func (ou *OnlineUser) IsFavoriteGame(gameId string) bool {
	return slices.Contains(ou.Favorites, gameId)
}

// AddFavoriteGame 添加到游戏收藏
func (ou *OnlineUser) AddFavoriteGame(gameId string) {
	found := ou.IsFavoriteGame(gameId)
	if !found {
		ou.Favorites = append(ou.Favorites, gameId)
	}
}

// RemoveFavoriteGame 从游戏收藏中移除
func (ou *OnlineUser) RemoveFavoriteGame(gameId string) {
	index := -1
	for i, id := range ou.Favorites {
		if id == gameId {
			index = i
			break
		}
	}
	if index != -1 {
		ou.Favorites = append(ou.Favorites[:index], ou.Favorites[index+1:]...)
	}
}

func (ou *OnlineUser) GetGameSaveData(gameId, key string) (string, error) {
	save, ok := ou.SaveData[gameId]
	if !ok {
		return "", apperrors.ErrUserGameSaveDataNotExisted
	}
	d, ok := save.SaveData[key]
	if !ok {
		return "", nil
	}
	return d, nil
}

func (ou *OnlineUser) SetGameSaveData(gameId, key, saveData string) {
	save, ok := ou.SaveData[gameId]
	if !ok {
		save = &UserSaveData{
			GameId:   gameId,
			SaveData: map[string]string{},
		}
		ou.SaveData[gameId] = save
	}
	save.SaveData[key] = saveData
}
