package dao

import (
	"dashfun_gamecenter/datasource/types"
)

func GetUserDao() types.UserDao {
	return daoImpl.GetUserDao()
}
func GetGameDao() types.GameDao {
	return daoImpl.GetGameDao()
}
func GetPaymentDao() types.PaymentDao {
	return daoImpl.GetPaymentDao()
}
func GetTaskDao() types.TaskDao {
	return daoImpl.GetTaskDao()
}
func GetTaskUserDao() types.TaskUserDao {
	return daoImpl.GetTaskUserDao()
}
func GetCoinDao() types.CoinDao {
	return daoImpl.GetCoinDao()
}
func GetCoinUserDao() types.CoinUserDao {
	return daoImpl.GetCoinUserDao()
}
func GetCoinRecordDao() types.CoinRecordDao {
	return daoImpl.GetCoinRecordDao()
}
func GetAdminUserDao() types.AdminUserDao { return daoImpl.GetAdminUserDao() }
func GetAdminUserLoginInfoDao() types.AdminUserLoginInfoDao {
	return daoImpl.GetAdminUserLoginInfoDao()
}
func GetSpinWheelDao() types.SpinWheelDao {
	return daoImpl.GetSpinWheelDao()
}
func GetSpinWheelUserDao() types.SpinWheelUserDao {
	return daoImpl.GetSpinWheelUserDao()
}

func GetUserSaveDataDao() types.DashFunUserSaveDataDao {
	return daoImpl.GetUserSaveDataDao()
}

func GetUserPlayRecordDao() types.DashFunUserPlayRecordDao {
	return daoImpl.GetUserPlayRecordDao()
}

func GetInvitedUserDao() types.InvitedUserDao { return daoImpl.GetInvitedUserDao() }

func GetRechargeDao() types.RechargeDao {
	return daoImpl.GetRechargeDao()
}
func GetLeaderboardBotDao() types.LeaderboardBotDao {
	return daoImpl.GetLeaderboardBotDao()
}

func GetAirdropDao() types.AirdropDao {
	return daoImpl.GetAirdropDao()
}

func GetUserProfileDao() types.UserProfileDao {
	return daoImpl.GetUserProfileDao()
}

func GetFishingPostDao() types.FishingPostDao {
	return daoImpl.GetFishingPostDao()
}

func GetFishingLeaderboardBotDao() types.FishingLeaderboardBotDao {
	return daoImpl.GetFishingLeaderboardBotDao()
}

func GetNolanDevPostDao() types.NolanDevPostDao {
	return daoImpl.GetNolanDevPostDao()
}

func GetNolanDevLeaderboardBotDao() types.NolanDevLeaderboardBotDao {
	return daoImpl.GetNolanDevLeaderboardBotDao()
}

func GetPricePredictDao() types.PricePredictDao {
	return daoImpl.GetPricePredictDao()
}

func GetExchangeDao() types.ExchangeDao {
	return daoImpl.GetExchangeDao()
}
