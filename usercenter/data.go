package usercenter

import (
	"dashfun_gamecenter/datasource/data"
	"sync"
	"time"
)

type OnlineUsers struct {
	// key = userId
	Users map[string]*data.OnlineUser
	//channel_id -> userId
	ChannelMap map[string]string
	sync.RWMutex
}

func newOnlineUsers() *OnlineUsers {
	return &OnlineUsers{
		Users:      make(map[string]*data.OnlineUser),
		ChannelMap: make(map[string]string),
	}
}

func (o *OnlineUsers) FindUserByChannelId(channelId string) *data.OnlineUser {
	o.RLock()
	defer o.RUnlock()

	userId, exist := o.ChannelMap[channelId]
	if !exist {
		return nil
	}
	user, exist := o.Users[userId]
	if !exist {
		return nil
	}
	return user
}

func (o *OnlineUsers) FindUser(userId string) *data.OnlineUser {
	o.RLock()
	defer o.RUnlock()

	user, exist := o.Users[userId]
	if !exist {
		return nil
	}
	return user
}

func (o *OnlineUsers) TGUserLogin(user *data.DashFunUser, tgInfo *data.TGInfo, playRecord []*data.PlayGameRecord) *data.OnlineUser {
	o.Lock()
	defer o.Unlock()

	u, e := o.Users[user.Id]
	if !e {
		u = data.NewOnlineUser(user, tgInfo, playRecord)
		o.Users[user.Id] = u
	}
	o.ChannelMap[user.ChannelId] = user.Id
	u.User.LoginTime = time.Now().UnixMilli()
	return u
}

func (o *OnlineUsers) UserLogout(user *data.DashFunUser) *data.OnlineUser {
	o.Lock()
	defer o.Unlock()

	u, e := o.Users[user.Id]
	if e {
		u.User.LogoffTime = time.Now().UnixMilli()
		delete(o.Users, user.Id)
		delete(o.ChannelMap, user.ChannelId)
	}
	return u
}
