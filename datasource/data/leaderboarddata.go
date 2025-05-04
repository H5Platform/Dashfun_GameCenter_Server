package data

type LeaderboardBotStatus int

const (
	LeaderboardBotStatus_Active    LeaderboardBotStatus = iota + 1 //活跃
	LeaderboardBotStatus_DoneToday                                 //今天已完成
)

type LeaderboardBotData struct {
	Id         string               `json:"id" bson:"_id"`                  //user id
	Name       string               `json:"name" bson:"name"`               //bot 名称
	Level      int                  `json:"level" bson:"level"`             //bot 级别，分6档，每一档的活跃程度不同
	Score      int64                `json:"score" bson:"score"`             //分数
	Rank       int                  `json:"rank" bson:"rank"`               //当前排名
	ActiveDays int                  `json:"active_days" bson:"active_days"` //活跃天数，从0开始
	ActiveTime int                  `json:"active_time" bson:"active_time"` //活跃时间，一天24小时之内随机一个时间，到点活跃
	ActiveDate string               `json:"active_date" bson:"active_date"` //记录最后一次活跃的日期，格式为YYYYMMDD UTC时间
	Status     LeaderboardBotStatus `json:"status" bson:"status"`           //bot状态
}

type LeaderboardScoreType string

const (
	LeaderboardScoreType_XPTotal LeaderboardScoreType = "dashfun_point_total" //总积分榜
	LeaderboardScoreType_XPDelta LeaderboardScoreType = "dashfun_point_delta" //增量积分榜
)

type LeaderboardPeriodType int32

const (
	LeaderboardPeriod_Forever LeaderboardPeriodType = iota + 1 //永久榜
	LeaderboardPeriod_Daily                                    //日榜
	LeaderboardPeriod_Weekly                                   //周榜
	LeaderboardPeriod_Monthly                                  //月榜
	LeaderboardPeriod_Yearly                                   //年榜
)

func LeaderboardPeriodTypeToString(periodType LeaderboardPeriodType) string {
	switch periodType {
	case LeaderboardPeriod_Forever:
		return "Forever"
	case LeaderboardPeriod_Daily:
		return "Daily"
	case LeaderboardPeriod_Weekly:
		return "Weekly"
	case LeaderboardPeriod_Monthly:
		return "Monthly"
	case LeaderboardPeriod_Yearly:
		return "Yearly"
	default:
		return "Unknown"
	}
}

type LeaderboardStatus int32 //排行榜状态

const (
	LeaderboardStatus_Active   LeaderboardStatus = iota + 1 //排行榜正常
	LeaderboardStatus_Reseting                              //排行榜重置中
)

type LeaderboardDefine struct {
	Id         string                `json:"id" bson:"_id"`                //排行榜ID
	Name       string                `json:"name" bson:"name"`             //排行榜名称
	GameId     string                `json:"game_id" bson:"game_id"`       //绑定的游戏ID，空或者DashFun表示DashFun平台
	PeriodType LeaderboardPeriodType `json:"type" bson:"type"`             //排行榜周期类型
	ScoreType  string                `json:"score_type" bson:"score_type"` //排行榜分数类型，由使用者定义，当上报分数给Leaderboard时，会更新所有匹配这个ScoreType的排行榜
	InitTime   int64                 `json:"init_time" bson:"init_time"`   //排行榜初始化时间，Unix时间戳，单位秒
	ResetTime  int64                 `json:"reset_time" bson:"reset_time"` //排行榜重置时间，Unix时间戳，单位秒，永久类型为0
	Status     LeaderboardStatus     `json:"status" bson:"status"`         //排行榜状态
}

type LeaderboardRankData struct {
	UserId string `json:"user_id" bson:"_id"` //用户ID
	Rank   int64  `json:"rank" bson:"rank"`   //用户排名
	Score  int64  `json:"score" bson:"score"` //用户分数
}

type LeaderboardHistory struct {
	LeaderboardId string `json:"leaderboard_id" bson:"leaderboard_id"` //排行榜ID
	UserId        string `json:"user_id" bson:"user_id"`               //用户ID
	Score         int64  `json:"score" bson:"score"`                   //用户分数
	Rank          int    `json:"rank" bson:"rank"`                     //用户排名
	StartTime     string `json:"start_time" bson:"start_time"`         //排行榜开始时间
}
