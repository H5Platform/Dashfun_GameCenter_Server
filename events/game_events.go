package events

import "dashfun_gamecenter/datasource/data"

type EventPlayerLevelUp struct {
	User  *data.DashFunUser
	Game  *data.DashFunGame
	Level int
}

var PlayerLevelUpEvents = NewEvent[*EventPlayerLevelUp]()
