package events

import "dashfun_gamecenter/datasource/data"

type EventUserEnterGame struct {
	User *data.DashFunUser
	Game *data.DashFunGame
}

var UserLoginEvents = NewEvent[*data.OnlineUser]()
var UserLogoffEvents = NewEvent[*data.OnlineUser]()
var UserEnterGameEvents = NewEvent[*EventUserEnterGame]()
