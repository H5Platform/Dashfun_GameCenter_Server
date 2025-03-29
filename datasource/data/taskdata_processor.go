package data

// TaskSaveDataPlayRandomGame PlayRandomGame类型的任务处理器保存的数据结构
type TaskSaveDataPlayRandomGame struct {
	Games []string `json:"games"` //玩过的游戏id
}

type TaskSaveDataFollowX struct {
	RandomCount int `json:"random_count"` //随机让用户进行check的次数
	CheckCount  int `json:"check_count"`  //已经check了的次数
}
