package data

type CoinData struct {
	Id          string            `json:"id" bson:"_id"`
	Name        string            `json:"name" bson:"name"`
	Symbol      string            `json:"symbol" bson:"symbol"`
	Desc        string            `json:"desc" bson:"desc"`
	BindGameId  string            `json:"bind_game_id" bson:"bind_game_id"` //绑定的游戏id，如果不填则绑定DashFun，填写则绑定制定游戏，一个游戏只能绑定一个coin
	CanWithdraw bool              `json:"can_withdraw" bson:"can_withdraw"` //是否可以提取
	MinWithdraw float32           `json:"min_withdraw" bson:"min_withdraw"` //最低提取金额
	ChainAddr   map[string]string `json:"chain_addr" bson:"chain_addr"`     //链上地址，chainName->address
	CreateTime  int64             `json:"-" bson:"create_time"`
}

type CoinUserData struct {
	UserId     string `json:"user_id" bson:"user_id"`
	CoinId     string `json:"coin_id" bson:"coin_id"`
	Amount     int32  `json:"amount" bson:"amount"`
	CreateTime int64  `json:"create_time" bson:"create_time"`
}

// CoinUserRecordData 用户coin变化记录
type CoinUserRecordData struct {
	UserId string `json:"user_id" bson:"user_id"`
	CoinId string `json:"coin_id" bson:"coin_id"`
	Change int32  `json:"change" bson:"change"`
	Reason string `json:"reason" bson:"reason"` //变化原因
	Info   string `json:"info" bson:"info"`     //变化信息
	Time   int64  `json:"time" bson:"time"`
}
