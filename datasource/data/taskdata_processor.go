package data

// TaskSaveDataPlayRandomGame PlayRandomGame类型的任务处理器保存的数据结构
type TaskSaveDataPlayRandomGame struct {
	Games []string `json:"games"` //玩过的游戏id
}
