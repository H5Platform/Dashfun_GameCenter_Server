package data

type DashFunUserSaveData struct {
	UserId string `json:"user_id" bson:"user_id"` //用户Id
	GameId string `json:"game_id" bson:"game_id"` //游戏Id
	Key    string `json:"key" bson:"key"`         //数据键值
	Data   string `json:"data" bson:"data"`       //保存的数据
}
