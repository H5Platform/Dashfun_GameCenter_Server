package events

import (
	"dashfun_gamecenter/datasource/data"
)

type EventUserEnterGame struct {
	User *data.DashFunUser
	Game *data.DashFunGame
}

type EventUserPayment struct {
	User    *data.DashFunUser
	Game    *data.DashFunGame
	Payment *data.DashFunPaymentData
}

type EventUserBindWallet struct {
	User    *data.DashFunUser
	Chain   string
	Address string
}

var UserLoginEvents = NewEvent[*data.OnlineUser]()
var UserLogoffEvents = NewEvent[*data.OnlineUser]()
var UserEnterGameEvents = NewEvent[*EventUserEnterGame]()

var UserPaymentEvents = NewEvent[*EventUserPayment]()
var UserBindAddressEvents = NewEvent[*EventUserBindWallet]()
