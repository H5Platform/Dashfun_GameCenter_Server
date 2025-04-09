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
