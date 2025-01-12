package taskcenter

import (
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

func (t *TaskCenter) onUserLogout(user *data.OnlineUser) {

}

func (t *TaskCenter) onUserPayment(evt *events.EventUserPayment) {
	user := evt.User
	payment := evt.Payment

	t.tasksLock.RLock()
	defer t.tasksLock.RUnlock()

	for _, task := range t.tasks {
		if task.Open && (isDashFunTask(task) || task.GameId == payment.GameId) {
			changed := false
			userData, err := t.GetTaskUserData(user.Id, task.Id)
			if err != nil {
				zap.S().Errorw("GetTaskUserData Error", "user", user.Id, "task", task.Id, "err", err)
				continue
			}

			switch task.Condition.Type {
			case data.TaskCondition_SpendTGStars:
				changed = t.taskRecordUserPayment(user, task, userData, payment, payment.GameId)
			}

			if changed {
				userData.Time = time.Now().UnixMilli()
				t.saveTaskUserData(userData)
			}
		}
	}

}

func (t *TaskCenter) onGameReportPlayerLevelUp(evt *events.EventPlayerLevelUp) {
	user := evt.User
	game := evt.Game

	t.tasksLock.RLock()
	defer t.tasksLock.RUnlock()

	for _, task := range t.tasks {
		if task.Open && (isDashFunTask(task) || task.GameId == game.Id) {
			changed := false
			userData, err := t.GetTaskUserData(user.Id, task.Id)
			if err != nil {
				zap.S().Errorw("GetTaskUserData Error", "user", user.Id, "task", task.Id, "err", err)
				continue
			}

			switch task.Condition.Type {
			case data.TaskCondition_LevelUp:
				changed = t.taskRecordPlayerLevelUp(user, task, userData, game.Id, evt.Level)
				break
			}

			if changed {
				userData.Time = time.Now().UnixMilli()
				t.saveTaskUserData(userData)
			}
		}
	}
}

// onUserEnterGameEvent 用户点击了Play按钮进入了游戏
func (t *TaskCenter) onUserEnterGameEvent(evt *events.EventUserEnterGame) {
	//t.processTasks(evt.User, evt.Game)
	user := evt.User
	game := evt.Game

	t.tasksLock.RLock()
	defer t.tasksLock.RUnlock()

	for _, task := range t.tasks {
		if task.Open && (isDashFunTask(task) || task.GameId == game.Id) {
			changed := false
			userData, err := t.GetTaskUserData(user.Id, task.Id)
			if err != nil {
				zap.S().Errorw("GetTaskUserData Error", "user", user.Id, "task", task.Id, "err", err)
				continue
			}

			switch task.Condition.Type {
			case data.TaskCondition_JoinTGChannel:
				//加入tg channel
				//@2024-11-07 改为用户手动点击验证了
				// changed = t.taskVerifyTGChannel(user, task, userData, game.Id)
				break
			case data.TaskCondition_PlayGame:
				//进行指定游戏
				changed = t.taskRecordPlayGame(user, task, userData, game.Id)
				break
			case data.TaskCondition_PlayRandomGame:
				//进行任意游戏
				changed = t.taskRecordPlayRandomGame(user, task, userData, game.Id)
				break
			case data.TaskCondition_LevelUp:
				break
			}

			if changed {
				userData.Time = time.Now().UnixMilli()
				t.saveTaskUserData(userData)
			}
		}
	}
}

func (t *TaskCenter) onUserBindAddress(evt *events.EventUserBindWallet) {
	user := evt.User

	t.tasksLock.RLock()
	defer t.tasksLock.RUnlock()

	for _, task := range t.tasks {
		if task.Open && (isDashFunTask(task)) { // || task.GameId == game.Id) {
			changed := false
			userData, err := t.GetTaskUserData(user.Id, task.Id)
			if err != nil {
				zap.S().Errorw("GetTaskUserData Error", "user", user.Id, "task", task.Id, "err", err)
				continue
			}

			switch task.Condition.Type {
			case data.TaskCondition_BindWallet:
				if task.Condition.Condition == evt.Chain {
					changed = t.taskVerifyWalletAddress(user, task, userData)
				}
				break
			}

			if changed {
				userData.Time = time.Now().UnixMilli()
				t.saveTaskUserData(userData)
			}
		}
	}
}
