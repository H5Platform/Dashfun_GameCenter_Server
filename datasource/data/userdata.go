package data

import initdata "github.com/telegram-mini-apps/init-data-golang"

type DashFunUserFrom int

const (
	DF_UserFrom_TG DashFunUserFrom = iota + 1 //telegram 用户
)

// DashFunUser user data db
type DashFunUser struct {
	Id          string          `json:"id" bson:"_id"`                    //全局userId
	ChannelId   string          `json:"channel_id" bson:"channel_id"`     //渠道方id
	DisplayName string          `json:"display_name" bson:"display_name"` //显示名称
	UserName    string          `json:"user_name" bson:"user_name"`       //用户名
	AvatarUrl   string          `json:"avatar_url" bson:"avatar_url"`     //avatar地址
	From        DashFunUserFrom `json:"from" bson:"from"`                 //用户来源
	CreateData  int64           `json:"create_data" bson:"create_data"`   //创建时间
	LoginTime   int64           `json:"login_time" bson:"login_time"`     //登录时间
	LogoffTime  int64           `json:"logoff_time" bson:"logoff_time"`   //登出时间
}

type TGInfo struct {
	AuthData string
	InitData *initdata.InitData
}

type OnlineUser struct {
	User   *DashFunUser
	TGInfo *TGInfo
}
