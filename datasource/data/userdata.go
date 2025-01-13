package data

import (
	"dashfun_gamecenter/apperrors"
	initdata "github.com/telegram-mini-apps/init-data-golang"
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

func NewOnlineUser(User *DashFunUser, TGInfo *TGInfo) *OnlineUser {
	return &OnlineUser{
		User:     User,
		TGInfo:   TGInfo,
		SaveData: map[string]*UserSaveData{},
	}
}

type OnlineUser struct {
	User   *DashFunUser
	TGInfo *TGInfo
	//用户在各个游戏中存储的数据
	//GameId -> *UserSaveData
	SaveData map[string]*UserSaveData
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
