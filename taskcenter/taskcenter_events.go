package taskcenter

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"go.uber.org/zap"
)

func (t *TaskCenter) onUserLogin(user *data.OnlineUser) {
	//用户登录，检查并读取用户的任务进度数据
	err := t.LoadAllTaskUserData(user.User.Id)
	if err != nil {
		zap.S().Errorw("LoadAllTaskUserData Error", "user", user.User.Id, "err", err)
	}
}

// onUserEnterGameEvent 用户点击了Play按钮进入了游戏
func (t *TaskCenter) onUserEnterGameEvent(evt *events.EventUserEnterGame) {

}
