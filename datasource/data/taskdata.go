package data

type DashFunTaskType int          //任务类型
type DashFunTaskConditionType int //任务条件类型
type DashFunTaskCategory int      //任务分类
type DashFunTaskRewardType int    //任务奖励类型
type DashFunTaskStatus int        //任务执行状态

const (
	TaskType_Normal DashFunTaskType = iota + 1 //普通类型任务，只能完成1次
	TaskType_Daily                             //每日任务，每日重置
	TaskType_2Days                             //2日任务，不知道用得上不...
)

const (
	TaskCondition_PlayRandomGame DashFunTaskConditionType = iota + 1 //订阅型任务，例如follow twitter, join tg channel...
	TaskCondition_PlayGame                                           //玩指定游戏指定次数，游戏id在task中指定
	TaskCondition_LevelUp                                            //在指定游戏中升级到指定等级
	TaskCondition_JoinTGChannel                                      //加入指定的tg channel
	TaskCondition_FollowX                                            //Follow X
	TaskCondition_SpendTGStars                                       //在TG中花费星星
)

const (
	TaskCategory_Challenges DashFunTaskCategory = iota + 1
	TaskCategory_Daily                          //每日
)

const (
	TaskRewardType_DashFunToken DashFunTaskRewardType = iota + 1 //奖励DashFunToken
	TaskRewardType_DashFunPoint                                  //奖励dashfun point，用来兑换链上token
)

func TaskRewardType2CoinName(rewardType DashFunTaskRewardType) string {
	switch rewardType {
	case TaskRewardType_DashFunToken:
		return "DashFunCoin"
	case TaskRewardType_DashFunPoint:
		return "DashFunCoin"
	default:
		return "DashFunCoin"
	}
}

const (
	TaskStatus_InProgress     DashFunTaskStatus = iota + 1 //任务正在进行中
	TaskStatus_Verify_Pending                              //任务需要验证
	TaskStatus_Completed
	TaskStatus_Claimed //任务奖励已领取
)

type DashFunTaskReward struct {
	RewardType DashFunTaskRewardType `json:"reward_type" bson:"reward_type"` //奖励类型
	Amount     float32               `json:"amount" bson:"amount"`           //奖励数量
}

type DashFunTaskCondition struct {
	Type      DashFunTaskConditionType `json:"type" bson:"type"`           //任务条件类型，不同类型任务完成方式不同
	Count     int                      `json:"count" bson:"count"`         //任务要求满足条件的次数
	Condition string                   `json:"condition" bson:"condition"` //任务条件，不同类型条件不同
	Link      string                   `json:"link" bson:"link"`           //任务条件相关链接
}

// DashFunTaskData 任务数据
type DashFunTaskData struct {
	Id         string               `json:"id" bson:"_id"`                  //任务ID
	Name       string               `json:"name" bson:"name"`               //任务名称
	Open       bool                 `json:"open" bson:"open"`               //任务是否开启
	GameId     string               `json:"game_id" bson:"game_id"`         //绑定游戏id，-1或""表示不限制游戏
	Type       DashFunTaskType      `json:"task_type" bson:"task_type"`     //任务类型
	Category   DashFunTaskCategory  `json:"category" bson:"category"`       //任务分类
	Condition  DashFunTaskCondition `json:"require" bson:"require"`         //任务条件
	Reward     DashFunTaskReward    `json:"reward" bson:"reward"`           //任务奖励
	CreateTime int64                `json:"create_time" bson:"create_time"` //任务创建时间
}

// DashFunTaskUserData 用户任务数据
type DashFunTaskUserData struct {
	UserId   string            `json:"user_id" bson:"user_id"`     //user id
	TaskId   string            `json:"task_id" bson:"task_id"`     //task id
	Progress int               `json:"progress" bson:"progress"`   //任务进度，max是任务对应的Condition.Count
	Status   DashFunTaskStatus `json:"status" bson:"status"`       //任务状态，0=in progress,1=
	SaveData string            `json:"save_data" bson:"save_data"` //任务进度相关数据
	Time     int64             `json:"time" bson:"time"`           //任务最新一次的进度变化时间
}

// UserTaskInfo 用户任务信息
// 存放用户当前可用的任务列表，以及任务对应的进度
type UserTaskInfo struct {
	Tasks    []*DashFunTaskData              `json:"tasks"`     //任务列表
	UserData map[string]*DashFunTaskUserData `json:"user_data"` //任务id对应的用户进度数据
}
