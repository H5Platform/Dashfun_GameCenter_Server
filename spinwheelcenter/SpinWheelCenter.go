package spinwheelcenter

import (
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/snowflake"
	"dashfun_gamecenter/utils"
	"errors"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var once sync.Once
var instance *SpinWheelCenter

type SpinWheelCenter struct {
	idGen *snowflake.Worker
}

func Get() *SpinWheelCenter {
	once.Do(func() {
		instance = &SpinWheelCenter{}
		instance.init()
	})
	return instance
}

func (s *SpinWheelCenter) newSpinWheelId() string {
	id := s.idGen.NextId()
	return strconv.FormatInt(id, 36)
}

func (s *SpinWheelCenter) init() {
	s.idGen = snowflake.Must(snowflake.GetWorker(data.WorkerSpinWheelId))
}

func (s *SpinWheelCenter) CreateWheelForGame(name, gameId string, rewards []data.SpinWheelReward) (*data.SpinWheelData, error) {
	id := s.newSpinWheelId()
	wheel, err := dao.GetSpinWheelDao().CreateSpinWheel(id, name, gameId, rewards)
	if err != nil {
		return nil, err
	}
	return wheel, nil
}

// GetSpinWheelForGame 获取游戏的转盘数据
// Deprecated: 目前弃用，不需要给指定游戏设置轮盘数据了 目前弃用，不需要给指定游戏设置轮盘数据了
func (s *SpinWheelCenter) GetSpinWheelForGame(gameId string) (*data.SpinWheelData, error) {
	return dao.GetSpinWheelDao().GetGameSpinWheel(gameId)
}

func (s *SpinWheelCenter) GetDashFunSpinWheelData() *data.SpinWheelData {
	cfg := config.GetConfig().SpinWheelCfg

	return &data.SpinWheelData{
		Id:      "DashFunSpinWheel",
		Name:    "DashFunSpinWheel",
		GameId:  "",
		Rewards: cfg.Rewards,
	}

}

func (s *SpinWheelCenter) GetSpinWheelUserData(userId, spinWheelId string) (*data.SpinWheelUserData, error) {
	ud := dao.GetSpinWheelUserDao()
	userData, err := ud.GetUserSpinWheelData(userId, spinWheelId)
	if err != nil {
		return nil, err
	}
	if userData == nil {
		userData = &data.SpinWheelUserData{
			UserId:      userId,
			SpinWheelId: spinWheelId,
			RewardIndex: 0,
			SpinTime:    0,
			Count:       0,
			Status:      data.SpinWheelUserStatus_Spin,
		}
		ud.SaveOrUpdate(userData)
	}

	if userData.Status == data.SpinWheelUserStatus_Claimed {
		//检查是否需要重置
		spinTime := time.UnixMilli(userData.SpinTime)
		needReset := utils.NeedReset(spinTime)
		//if config.IsProd() {
		//	needReset = !utils.IsSameDay(spinTime, nowTime)
		//} else {
		//	//测试服务器，1分钟的cd
		//	needReset = (nowTime.UnixMilli() - spinTime.UnixMilli()) > 1000*60
		//}

		if needReset {
			//重置
			userData.Status = data.SpinWheelUserStatus_Spin
			userData.SpinTime = 0
			userData.RewardIndex = 0
			userData.Count = 0
			ud.SaveOrUpdate(userData)
		}
	}
	return userData, nil
}

// UserSpinWheel 用户转轮盘，gameId目前没用了，统一抽DashFun的轮盘
func (s *SpinWheelCenter) UserSpinWheel(userId, gameId string) (*data.SpinWheelUserData, error) {
	spinData := s.GetDashFunSpinWheelData()
	userData, err := s.GetSpinWheelUserData(userId, spinData.Id)
	if err != nil {
		return nil, err
	}

	if userData.Status == data.SpinWheelUserStatus_CanClaim {
		return nil, errors.New("invalid spin status")
	}

	if userData.Count >= len(config.GetConfig().SpinWheelCfg.TicketsNeeded) {
		return nil, errors.New("invalid spin count")
	}

	ticketsNeeded := config.GetConfig().SpinWheelCfg.TicketsNeeded[userData.Count]

	ticket := coincenter.Get().GetDashFunTicket()
	userTicket := coincenter.Get().GetCoinUserData(userId, ticket.Id)
	if userTicket.Amount < int32(ticketsNeeded) {
		return nil, errors.New("not enough tickets")
	}

	hit := s.RandomSpinWheel()

	if hit == nil {
		return nil, errors.New("invalid spin data")
	}

	//扣除票数
	_, err = coincenter.Get().DecUserCoinAmount(userId, ticket.Id, int32(ticketsNeeded), "SpinWheel", strconv.Itoa(userData.Count))

	userData.RewardIndex = hit.RewardIndex
	userData.RewardValue = int(float64(hit.RewardValue) * (1 + 0.2*rand.Float64()))
	userData.SpinTime = time.Now().UnixMilli()
	userData.Status = data.SpinWheelUserStatus_CanClaim
	userData.Count++
	dao.GetSpinWheelUserDao().SaveOrUpdate(userData)

	return userData, nil

}

func (s *SpinWheelCenter) RandomSpinWheel() (hit *data.SpinWheelReward) {
	spinData := s.GetDashFunSpinWheelData()
	total := 0
	for _, reward := range spinData.Rewards {
		total += reward.Weight
	}

	v := rand.Intn(total)
	total = 0
	for _, reward := range spinData.Rewards {
		total += reward.Weight
		if total > v {
			hit = &reward
			break
		}
	}
	return
}

func (s *SpinWheelCenter) UserClaimReward(userId, gameId string) (*data.SpinWheelUserData, error) {
	//spinWheelData, err := s.GetSpinWheelForGame(gameId)
	//if err != nil {
	//	return nil, err
	//}

	spinWheelData := s.GetDashFunSpinWheelData()
	userData, err := s.GetSpinWheelUserData(userId, spinWheelData.Id)
	if err != nil {
		return nil, err
	}

	if userData.Status != data.SpinWheelUserStatus_CanClaim {
		return nil, errors.New("invalid spin status")
	}

	reward := s.findSpinWheelRewardByIndex(spinWheelData, userData.RewardIndex)
	if reward == nil {
		return nil, errors.New("invalid spin reward")
	}

	var coin *data.CoinData
	var exist = false

	switch reward.RewardType {
	case data.SpinWheelReward_DashFunPoint:
		coin = coincenter.Get().GetDashFunXp()
		exist = true
	case data.SpinWheelReward_GamePoint:
		coin, exist = coincenter.Get().GetCoinByGame(gameId)
	}
	if !exist {
		return nil, errors.New("invalid spin reward type")
	}

	_, err = coincenter.Get().AddUserCoinAmount(userId, coin.Id, int32(reward.RewardValue), "SpinWheelReward", "")
	if err != nil {
		return nil, err
	}

	userData.Status = data.SpinWheelUserStatus_Claimed
	dao.GetSpinWheelUserDao().SaveOrUpdate(userData)

	return userData, nil
}

func (s *SpinWheelCenter) findSpinWheelRewardByIndex(spinwheelData *data.SpinWheelData, index int) *data.SpinWheelReward {
	for _, reward := range spinwheelData.Rewards {
		if reward.RewardIndex == index {
			return &reward
		}
	}
	return nil
}
