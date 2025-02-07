package data

// DashFunUserPlayRecord 用户玩游戏的记录
type DashFunUserPlayRecord struct {
	UserId  string            `json:"user_id" bson:"_id"` //用户Id
	Records []*PlayGameRecord `json:"records" bson:"records"`
}

type PlayGameRecord struct {
	GameId   string `json:"game_id" bson:"game_id"`
	PlayTime int64  `json:"play_time" bson:"play_time"`
}
