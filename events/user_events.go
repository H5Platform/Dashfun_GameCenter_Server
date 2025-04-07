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

type EventUserRecharge struct {
	User     *data.DashFunUser
	Game     *data.DashFunGame
	Recharge *data.DashFunRechargeData
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

type UserLeaderboardEvent struct {
	Id     string //排行榜Id，因为还没没有区分排行榜，等有多个的时候再改，目前只是空串
	UserId string
	Rank   int64
	Score  float64
}

var UserLoginEvents = NewEvent[*data.OnlineUser]()
var UserLogoffEvents = NewEvent[*data.OnlineUser]()
var UserEnterGameEvents = NewEvent[*EventUserEnterGame]()
var UserReferrerEvents = NewEvent[*UserReferrerEvent]()
var UserReferSuccessEvents = NewEvent[*data.InvitedUserData]()

var UserPaymentEvents = NewEvent[*EventUserPayment]()
var UserTGPaymentEvents = NewEvent[*EventUserPayment]()
var UserBindAddressEvents = NewEvent[*EventUserBindWallet]()
var UserRechargeEvents = NewEvent[*EventUserRecharge]()
var UserLeaderboardEvents = NewEvent[*UserLeaderboardEvent]()
