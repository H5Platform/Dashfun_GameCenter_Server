package types

import (
	"context"
	"dashfun_gamecenter/datasource/data"
)

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
	GetSpinWheelDao() SpinWheelDao
	GetSpinWheelUserDao() SpinWheelUserDao
	GetUserSaveDataDao() DashFunUserSaveDataDao
	GetUserPlayRecordDao() DashFunUserPlayRecordDao
	GetInvitedUserDao() InvitedUserDao
	GetRechargeDao() RechargeDao
	GetLeaderboardBotDao() LeaderboardBotDao
}

type UserDao interface {
	GetUserById(userId string) (*data.DashFunUser, error)
	GetUserByChannelId(channelId string) (*data.DashFunUser, error)
	SaveOrUpdate(user *data.DashFunUser) (*data.DashFunUser, error)
}

type GameDao interface {
	GetAllGames(status data.DashFunGameStatus) ([]*data.DashFunGame, error)
	GetGameById(gameId string) (*data.DashFunGame, error)
	SaveOrUpdate(game *data.DashFunGame) (*data.DashFunGame, error)
	GetGameByName(gameName string) (*data.DashFunGame, error)
	FindGames(keyword string, genre, flags []int, status data.DashFunGameStatus, size, page int64) (games []*data.DashFunGame, totalPages int, err error)
	FindGameList(listType data.GameListType, count int) (games []*data.DashFunGame, err error)
}

type PaymentDao interface {
	FindPaymentById(id string) (*data.DashFunPaymentData, error)
	SaveOrUpdate(game *data.DashFunPaymentData) (*data.DashFunPaymentData, error)
	CreatePayment(id, userId, gameId, paymentId, title, desc, payload string, currency data.PaymentCurrency, from data.PaymentFrom, price int, extraData string) (*data.DashFunPaymentData, error)
}

type RechargeDao interface {
	SaveOrUpdate(recharge *data.DashFunRechargeData) (*data.DashFunRechargeData, error)
	FindRechargeById(rechargeId string) (*data.DashFunRechargeData, error)
	CreateRecharge(id, userId string, from data.RechargeFrom, gameId string, price int, priceType data.RechargePlatformOptionPriceType, diamond int, payload, message string, createAt int64) (*data.DashFunRechargeData, error)
	GetOrdersByStatus(status data.RechargeStatus) ([]*data.DashFunRechargeData, error)
}

type TaskDao interface {
	FindTaskById(id string) (*data.DashFunTaskData, error)
	FindTaskByName(name string) (*data.DashFunTaskData, error)
	FindAllTasks() []*data.DashFunTaskData
	SearchTask(gameId, keyword string, size, page int64) (tasks []*data.DashFunTaskData, totalPages int, err error)
	FindTaskByGameId(gameId string) ([]*data.DashFunTaskData, error)
	SaveOrUpdate(task *data.DashFunTaskData) (*data.DashFunTaskData, error)
	CreateTask(id, name, gameId string, showInGame bool, priority int, taskType data.DashFunTaskType, category data.DashFunTaskCategory, condition data.DashFunTaskCondition, rewards []data.DashFunTaskReward) (*data.DashFunTaskData, error)
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
	FindCoinByGameId(gameId string) *data.CoinData
	GetAllCoins() ([]*data.CoinData, error)
	CreateCoin(id, name, symbol, desc, bindGameId string, canWithdraw bool, minWithdraw float32, chainAddr map[string]string) (*data.CoinData, error)
}

type CoinUserDao interface {
	SaveOrUpdate(user *data.CoinUserData) (*data.CoinUserData, error)
	GetAllUserCoins(userId string) ([]*data.CoinUserData, error)
	GetCoinUserDataCursor(coinId string, batchSize int32) (Cursor[data.CoinUserData], error)
}

type CoinRecordDao interface {
	AddRecord(user *data.CoinUserRecordData) (*data.CoinUserRecordData, error)
	GetAllUserCoinRecords(userId, coinId string) ([]*data.CoinUserRecordData, error)
}

type SpinWheelDao interface {
	CreateSpinWheel(id, name, gameId string, rewards []data.SpinWheelReward) (*data.SpinWheelData, error)
	GetGameSpinWheel(gameId string) (*data.SpinWheelData, error)
	GetSpinWheelById(spinWheelId string) (*data.SpinWheelData, error)
}

type SpinWheelUserDao interface {
	SaveOrUpdate(userData *data.SpinWheelUserData) (*data.SpinWheelUserData, error)
	GetUserSpinWheelData(userId, gameId string) (*data.SpinWheelUserData, error)
}

type DashFunUserSaveDataDao interface {
	SaveOrUpdate(user *data.DashFunUserSaveData) (*data.DashFunUserSaveData, error)
	GetUserSaveData(userId, gameId, key string) (*data.DashFunUserSaveData, error)
}

type DashFunUserPlayRecordDao interface {
	SaveOrUpdate(user *data.DashFunUserPlayRecord) (*data.DashFunUserPlayRecord, error)
	GetUserPlayRecord(userId string) (*data.DashFunUserPlayRecord, error)
}

type InvitedUserDao interface {
	FindInvitedByUserId(userId string) ([]*data.InvitedUserData, error)
	FindInvitedByInvitedUserId(userId string) ([]*data.InvitedUserData, error)
	SaveOrUpdate(user *data.InvitedUserData) (*data.InvitedUserData, error)
}

type Cursor[T any] interface {
	Next(ctx context.Context) bool
	Data() (*T, error)
	Close()
}

type LeaderboardBotDao interface {
	SaveOrUpdate(bot *data.LeaderboardBotData) (*data.LeaderboardBotData, error)
	LoadAllBots() ([]*data.LeaderboardBotData, error)
}
