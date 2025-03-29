package data

type InvitedStatus int
type InvitedType int

const (
	InvitedStatus_Login   InvitedStatus = iota + 1 //邀请的用户已登陆
	InvitedStatus_Success                          //邀请的用户已经复合积分条件
)

const (
	InvitedType_NewUser   InvitedType = iota + 1 //新用户
	InvitedType_OldUser                          //老用户
	InvitedType_SleepUser                        //非活跃用户(90天未登录)
)

type InvitedUserInfo struct {
	Username    string `json:"username" bson:"username"`
	DisplayName string `json:"display_name" bson:"display_name"`
	Avatar      string `json:"avatar" bson:"avatar"`
}

type InvitedUserData struct {
	UserId          string          `json:"user_id" bson:"user_id"`                 //user id
	InvitedUserId   string          `json:"invited_user_id" bson:"invited_user_id"` //邀请的用户ID
	InvitedUserName InvitedUserInfo `json:"invited_user_name" bson:"invited_user_name"`
	InvitedStatus   InvitedStatus   `json:"invited_status" bson:"invited_status"` //邀请状态
	InvitedType     InvitedType     `json:"invited_type" bson:"invited_type"`     //邀请类型
	InvitedTime     int64           `json:"invited_time" bson:"invited_time"`
}
