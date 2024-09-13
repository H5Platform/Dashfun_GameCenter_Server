package types

import "dashfun_gamecenter/datasource/data"

type DaoImpl interface {
	GetUserDao() UserDao
	GetGameDao() GameDao
	GetPaymentDao() PaymentDao
	GetTaskDao() TaskDao
	GetTaskUserDao() TaskUserDao
	GetCoinDao() CoinDao
	GetCoinUserDao() CoinUserDao
	GetCoinRecordDao() CoinRecordDao
	GetAdminUserDao() AdminUserDao
	GetAdminUserLoginInfoDao() AdminUserLoginInfoDao
}

type UserDao interface {
	GetUserById(userId string) (*data.DashFunUser, error)
	GetUserByChannelId(channelId string) (*data.DashFunUser, error)
	SaveOrUpdate(user *data.DashFunUser) (*data.DashFunUser, error)
}

type GameDao interface {
	GetGameById(gameId string) (*data.DashFunGame, error)
	SaveOrUpdate(game *data.DashFunGame) (*data.DashFunGame, error)
	GetGameByName(gameName string) (*data.DashFunGame, error)
}

type PaymentDao interface {
	FindPaymentById(id string) (*data.DashFunPaymentData, error)
	SaveOrUpdate(game *data.DashFunPaymentData) (*data.DashFunPaymentData, error)
	CreatePayment(id, userId, gameId, paymentId, title, desc, payload, currency string, from data.PaymentFrom, price int, extraData string) (*data.DashFunPaymentData, error)
}

type TaskDao interface {
	FindTaskById(id string) (*data.DashFunTaskData, error)
	FindTaskByName(name string) (*data.DashFunTaskData, error)
	FindAllTasks() []*data.DashFunTaskData
	SaveOrUpdate(task *data.DashFunTaskData) (*data.DashFunTaskData, error)
	CreateTask(id, name, gameId string, taskType data.DashFunTaskType, category data.DashFunTaskCategory, condition data.DashFunTaskCondition, reward data.DashFunTaskReward) (*data.DashFunTaskData, error)
}

type TaskUserDao interface {
	FindTaskUserData(taskId string, userId string) (*data.DashFunTaskUserData, error)
	FindAllTaskUserData(taskId string) ([]*data.DashFunTaskUserData, error)
	SaveOrUpdate(user *data.DashFunTaskUserData) (*data.DashFunTaskUserData, error)
}

type CoinDao interface {
	SaveOrUpdate(task *data.CoinData) (*data.CoinData, error)
	FindCoinById(coinId string) (*data.CoinData, error)
	FindCoinByName(name string) (*data.CoinData, error)
	GetAllCoins() ([]*data.CoinData, error)
	CreateCoin(id, name, symbol, desc string, canWithdraw bool, minWithdraw float32, chainAddr map[string]string) (*data.CoinData, error)
}

type CoinUserDao interface {
	SaveOrUpdate(user *data.CoinUserData) (*data.CoinUserData, error)
	GetAllUserCoins(userId string) ([]*data.CoinUserData, error)
}

type CoinRecordDao interface {
	AddRecord(user *data.CoinUserRecordData) (*data.CoinUserRecordData, error)
	GetAllUserCoinRecords(userId, coinId string) ([]*data.CoinUserRecordData, error)
}
