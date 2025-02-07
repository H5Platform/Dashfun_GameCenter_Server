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
