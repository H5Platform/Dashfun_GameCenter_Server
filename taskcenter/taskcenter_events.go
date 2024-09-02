package taskcenter

import (
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"go.uber.org/zap"
	"time"
)

func (t *TaskCenter) onUserLogin(user *data.OnlineUser) {
	//用户登录，检查并读取用户的任务进度数据
	_, err := t.loadAllTaskUserData(user.User.Id)
	if err != nil {
		zap.S().Errorw("LoadAllTaskUserData Error", "user", user.User.Id, "err", err)
	}
}

// onUserEnterGameEvent 用户点击了Play按钮进入了游戏
func (t *TaskCenter) onUserEnterGameEvent(evt *events.EventUserEnterGame) {
	t.processTasks(evt.User, evt.Game)
}

func (t *TaskCenter) processTasks(user *data.DashFunUser, game *data.DashFunGame) {
	for _, task := range t.tasks {
		if task.Open && (task.GameId == "" || task.GameId == "-1" || task.GameId == game.Id) {
			changed := false
			userData, err := t.GetTaskUserData(user.Id, task.Id)
			if err != nil {
				zap.S().Errorw("GetTaskUserData Error", "user", user.Id, "task", task.Id, "err", err)
				continue
			}

			switch task.Condition.Type {
			case data.TaskCondition_JoinTGChannel:
				//加入tg channel

				break
			case data.TaskCondition_PlayGame:
				//进行指定游戏
				changed = t.taskProcessorPlayGame(user, task, userData, game.Id)
				break
			case data.TaskCondition_PlayRandomGame:
				//进行任意游戏
				changed = t.taskProcessorPlayRandomGame(user, task, userData, game.Id)
				break
			case data.TaskCondition_LevelUp:
				break
			}

			if changed {
				userData.Time = time.Now().UnixMilli()
				dao.GetTaskUserDao().SaveOrUpdate(userData)
			}
		}
	}
}
