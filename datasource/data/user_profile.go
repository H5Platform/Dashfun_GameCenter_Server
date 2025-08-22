package data

type UserProfileData struct {
	UserId   string `bson:"_id" json:"userId"`        // 用户ID
	Nickname string `bson:"nickname" json:"nickname"` // 用户昵称，唯一
	Avatar   string `bson:"avatar" json:"avatar"`     // 用户头像，只记录是否设置了头像
}
