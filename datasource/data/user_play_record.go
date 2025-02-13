package data

// DashFunUserPlayRecord 用户玩过的游戏记录和收藏记录
type DashFunUserPlayRecord struct {
	UserId    string            `json:"user_id" bson:"_id"` //用户Id
	Records   []*PlayGameRecord `json:"records" bson:"records"`
	Favorites []string          `json:"favorites" bson:"favorites"`
}

type PlayGameRecord struct {
	GameId   string `json:"game_id" bson:"game_id"`
	PlayTime int64  `json:"play_time" bson:"play_time"`
}
