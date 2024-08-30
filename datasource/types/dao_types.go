package types

import "dashfun_gamecenter/datasource/data"

type DaoImpl interface {
	GetUserDao() UserDao
	GetGameDao() GameDao
	GetPaymentDao() PaymentDao
	GetTaskDao() TaskDao
	GetTaskUserDao() TaskUserDao
}

type UserDao interface {
	GetUserById(userId string) (*data.DashFunUser, error)
	GetUserByChannelId(channelId string) (*data.DashFunUser, error)
	SaveOrUpdate(user *data.DashFunUser) (*data.DashFunUser, error)
}

type GameDao interface {
	GetGameById(gameId string) (*data.DashFunGame, error)
	SaveOrUpdate(game *data.DashFunGame) (*data.DashFunGame, error)
}

type PaymentDao interface {
	FindPaymentById(id string) (*data.DashFunPaymentData, error)
	SaveOrUpdate(game *data.DashFunPaymentData) (*data.DashFunPaymentData, error)
	CreatePayment(id, userId, gameId, paymentId, title, desc, payload, currency string, from data.PaymentFrom, price int, extraData string) (*data.DashFunPaymentData, error)
}

type TaskDao interface {
	FindTaskById(id string) (*data.DashFunTaskData, error)
	FindAllTasks() []*data.DashFunTaskData
	SaveOrUpdate(task *data.DashFunTaskData) (*data.DashFunTaskData, error)
	CreateTask(id, name, gameId string, taskType data.DashFunTaskType, category data.DashFunTaskCategory, condition data.DashFunTaskCondition, reward data.DashFunTaskReward) (*data.DashFunTaskData, error)
}

type TaskUserDao interface {
	FindTaskUserData(taskId string, userId string) (*data.DashFunTaskUserData, error)
	FindAllTaskUserData(taskId string) ([]*data.DashFunTaskUserData, error)
	SaveOrUpdate(user *data.DashFunTaskUserData) (*data.DashFunTaskUserData, error)
}
