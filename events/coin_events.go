package events

import "dashfun_gamecenter/datasource/data"

type UserCoinChangedEvent struct {
	UserId        string
	Coin          *data.CoinData
	UserData      *data.CoinUserData
	ChangedAmount int32 // 变化的数量，>0表示增加 <0表示减少
}

var UserCoinChangedEvents = NewEvent[UserCoinChangedEvent]()
