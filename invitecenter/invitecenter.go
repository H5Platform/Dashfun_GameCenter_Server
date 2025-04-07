package invitecenter

import (
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"go.uber.org/zap"
	"sync"
	"time"
)

var onceInviteCenter sync.Once
var instInviteCenter *InviteCenter

// InviteCenter 邀请中心
// 每个用户可邀请同一个用于一次
// 邀请成功后，被邀请人达到指定分数后，邀请人可获得积分
// 邀请人获得的积分和被邀请人的状态有关，按照新用户、老用户，给与不同的积分
type InviteCenter struct {
}

func Get() *InviteCenter {
	onceInviteCenter.Do(func() {
		instInviteCenter = &InviteCenter{}
		instInviteCenter.init()
	})
	return instInviteCenter
}

func (i *InviteCenter) init() {
	events.UserReferrerEvents.On(i.onUserReferrer)
	events.UserCoinChangedEvents.On(i.onUserCoinChanged)
	//events.UserPointChangedEvents.On(i.onUserPointChanged)
}

func (i *InviteCenter) onUserReferrer(evt *events.UserReferrerEvent) {
	//用户邀请成功
	d := dao.GetInvitedUserDao()
	var invitedType data.InvitedType
	//目前没有实现90天未登录的sleep用户邀请
	if evt.IsNewCreate {
		invitedType = data.InvitedType_NewUser
	} else {
		invitedType = data.InvitedType_OldUser
	}

	invited := &data.InvitedUserData{
		UserId:        evt.Referrer.Id,
		InvitedUserId: evt.User.Id,
		InvitedUserName: data.InvitedUserInfo{
			Username:    evt.User.UserName,
			DisplayName: evt.User.DisplayName,
			Avatar:      evt.User.AvatarUrl,
		},
		InvitedStatus: data.InvitedStatus_Login,
		InvitedType:   invitedType,
		InvitedTime:   time.Now().UnixMilli(),
	}
	d.SaveOrUpdate(invited)
}

// GetInvitedUsers 获取指定用邀请的用户列表
func (i *InviteCenter) GetInvitedUsers(userId string) ([]*data.InvitedUserData, error) {
	d := dao.GetInvitedUserDao()
	users, err := d.FindInvitedByUserId(userId)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (i *InviteCenter) onUserCoinChanged(evt events.UserCoinChangedEvent) {
	if evt.Coin.Name != "DashFunPoint" || evt.ChangedAmount < 0 {
		return
	}

	d := dao.GetInvitedUserDao()
	uds, err := d.FindInvitedByInvitedUserId(evt.UserId)
	if err != nil {
		zap.S().Errorw("InviteCenter.onUserPointChanged FindInvitedByInvitedUserId failed", "error", err.Error())
		return
	}
	for _, ud := range uds {
		if ud != nil && ud.InvitedStatus == data.InvitedStatus_Login {
			var reward *config.RewardPoint = nil
			for _, r := range config.GetConfig().InviteCfg.PointReward {
				if r.InviteUserType == int(ud.InvitedType) {
					{
						reward = &r
						break
					}
				}
			}
			if reward != nil && evt.UserData.Amount >= int32(config.GetConfig().InviteCfg.PointRequired) {
				//邀请成功
				ud.InvitedStatus = data.InvitedStatus_Success
				d.SaveOrUpdate(ud)
				//给用户加奖励

				xp, _ := coincenter.Get().GetCoinByName("DashFunPoint")
				coin, _ := coincenter.Get().GetCoinByName("DashFunCoin")

				if reward.RewardPoint > 0 {
					_, err := coincenter.Get().AddUserCoinAmount(ud.UserId, xp.Id, int32(reward.RewardPoint), "InviteReward", "")
					if err != nil {
						zap.S().Errorw("InviteCenter.onUserCoinChanged AddUserCoinAmount failed", "error", err.Error())
					}
				}
				if reward.RewardCoin > 0 {
					_, err = coincenter.Get().AddUserCoinAmount(ud.UserId, coin.Id, int32(reward.RewardCoin), "InviteReward", "")
					if err != nil {
						zap.S().Errorw("InviteCenter.onUserCoinChanged AddUserCoinAmount failed", "error", err.Error())
					}
				}
				events.UserReferSuccessEvents.Emit(ud)
			}
		}
	}
}
