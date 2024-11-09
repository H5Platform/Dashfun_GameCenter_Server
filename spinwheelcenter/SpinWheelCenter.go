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

func (s *SpinWheelCenter) GetSpinWheelForGame(gameId string) (*data.SpinWheelData, error) {
	return dao.GetSpinWheelDao().GetGameSpinWheel(gameId)
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
			Status:      data.SpinWheelUserStatus_Spin,
		}
		ud.SaveOrUpdate(userData)
	}

	if userData.Status == data.SpinWheelUserStatus_Claimed {
		//检查是否需要重置
		spinTime := time.UnixMilli(userData.SpinTime)
		nowTime := time.Now()

		needReset := false
		if config.IsProd() {
			needReset = !utils.IsSameDay(spinTime, nowTime)
		} else {
			//测试服务器，1分钟的cd
			needReset = (nowTime.UnixMilli() - spinTime.UnixMilli()) > 1000*60
		}

		if needReset {
			//重置
			userData.Status = data.SpinWheelUserStatus_Spin
			userData.SpinTime = 0
			userData.RewardIndex = 0
			ud.SaveOrUpdate(userData)
		}
	}
	return userData, nil
}

func (s *SpinWheelCenter) UserSpinWheel(userId, gameId string) (*data.SpinWheelUserData, error) {
	spinData, err := s.GetSpinWheelForGame(gameId)
	if err != nil {
		return nil, err
	}
	if spinData == nil {
		return nil, errors.New("SpinWheel Data Not Found")
	}
	userData, err := s.GetSpinWheelUserData(userId, spinData.Id)
	if err != nil {
		return nil, err
	}
	if userData.Status != data.SpinWheelUserStatus_Spin {
		return nil, errors.New("invalid spin status")
	}

	total := 0
	for _, reward := range spinData.Rewards {
		total += reward.Weight
	}

	v := rand.Intn(total)
	total = 0
	var hit *data.SpinWheelReward = nil
	for _, reward := range spinData.Rewards {
		total += reward.Weight
		if total > v {
			hit = &reward
			break
		}
	}

	if hit == nil {
		return nil, errors.New("invalid spin data")
	}

	userData.RewardIndex = hit.RewardIndex
	userData.SpinTime = time.Now().UnixMilli()
	userData.Status = data.SpinWheelUserStatus_CanClaim
	dao.GetSpinWheelUserDao().SaveOrUpdate(userData)

	return userData, nil

}

func (s *SpinWheelCenter) UserClaimReward(userId, gameId string) (*data.SpinWheelReward, error) {
	spinWheelData, err := s.GetSpinWheelForGame(gameId)
	if err != nil {
		return nil, err
	}
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
	case data.SpinWheelReward_GamePoint:
		coin, exist = coincenter.Get().GetCoinByGame(gameId)
	}
	if !exist {
		return nil, errors.New("invalid spin reward type")
	}

	_, err = coincenter.Get().AddUserCoinAmount(userId, coin.Id, float32(reward.RewardValue))
	if err != nil {
		return nil, err
	}

	userData.RewardIndex = 0
	userData.Status = data.SpinWheelUserStatus_Claimed
	dao.GetSpinWheelUserDao().SaveOrUpdate(userData)

	return reward, nil

}

func (s *SpinWheelCenter) findSpinWheelRewardByIndex(spinwheelData *data.SpinWheelData, index int) *data.SpinWheelReward {
	for _, reward := range spinwheelData.Rewards {
		if reward.RewardIndex == index {
			return &reward
		}
	}
	return nil
}
