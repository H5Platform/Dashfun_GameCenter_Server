package taskcenter

import (
	"dashfun_gamecenter/datasource/data"
	"encoding/json"
	"go.uber.org/zap"
)

// taskProcessorPlayGame
func (t *TaskCenter) taskProcessorPlayGame(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData, gameId string) bool {
	//玩指定游戏
	ret := false
	if task.Condition.Type == data.TaskCondition_PlayGame {
		if task.GameId == gameId && userData.Status == data.TaskStatus_InProgress {
			if userData.Progress < task.Condition.Count {
				userData.Progress = userData.Progress + 1
				ret = true
			}
			if userData.Progress >= task.Condition.Count {
				userData.Status = data.TaskStatus_Completed
				ret = true
			}
		}
	}
	return ret
}

// taskProcessorPlayRandomGame
func (t *TaskCenter) taskProcessorPlayRandomGame(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData, gameId string) bool {
	//玩指定游戏
	if task.Condition.Type == data.TaskCondition_PlayGame {
		var save = &data.TaskSaveDataPlayRandomGame{}
		if userData.SaveData != "" {
			err := json.Unmarshal([]byte(userData.SaveData), save)
			if err != nil {
				zap.S().Errorw("get task save data error", "err", err.Error(), "user", user.Id, "task", task.Id, "game", task.GameId, "savedata", userData.SaveData)
			}
		}

		for _, gid := range save.Games {
			if gid == gameId {
				//已经记录过这个游戏了
				return false
			}
		}

		ret := false

		if userData.Progress < task.Condition.Count {
			save.Games = append(save.Games, gameId)
			userData.Progress = userData.Progress + 1
			bytes, err := json.Marshal(save)
			if err != nil {
				zap.S().Errorw("set task save data error", "err", err.Error(), "user", user.Id, "task", task.Id, "game", task.GameId, "savedata", save)
			} else {
				userData.SaveData = string(bytes)
			}
			ret = true
		}
		if userData.Progress >= task.Condition.Count {
			userData.Status = data.TaskStatus_Completed
			ret = true
		}

		return ret
	}
	return false
}
