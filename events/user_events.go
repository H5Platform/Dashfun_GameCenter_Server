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

type UserReferrerEvent struct {
	User        *data.DashFunUser //用户
	Referrer    *data.DashFunUser //推荐人
	IsNewCreate bool              //是否是新创建用户
}

var UserLoginEvents = NewEvent[*data.OnlineUser]()
var UserLogoffEvents = NewEvent[*data.OnlineUser]()
var UserEnterGameEvents = NewEvent[*EventUserEnterGame]()
var UserReferrerEvents = NewEvent[*UserReferrerEvent]()
var UserReferSuccessEvents = NewEvent[*data.InvitedUserData]()

var UserPaymentEvents = NewEvent[*EventUserPayment]()
var UserTGPaymentEvents = NewEvent[*EventUserPayment]()
var UserBindAddressEvents = NewEvent[*EventUserBindWallet]()
