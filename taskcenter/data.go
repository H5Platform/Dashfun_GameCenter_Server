package taskcenter

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/utils"
	"sync"
)

type TaskUserDataList struct {
	sync.RWMutex
	// userId -> TasksUserData
	usersTaskData map[string]*TasksUserData
}

type TasksUserData struct {
	userId       string
	taskDataList utils.List[*data.DashFunTaskUserData]
}

func NewTaskUserDataList() *TaskUserDataList {
	return &TaskUserDataList{
		usersTaskData: make(map[string]*TasksUserData),
	}
}

func newTaskUserData(userId string) *TasksUserData {
	return &TasksUserData{
		userId:       userId,
		taskDataList: utils.NewList[*data.DashFunTaskUserData](),
	}
}

func (t *TaskUserDataList) HasRecord(userId string) bool {
	t.RLock()
	defer t.RUnlock()
	_, ok := t.usersTaskData[userId]
	return ok
}

// getUserDataList 获取用户的任务进度数据，如果不存在则创建
// 注意调用前需要lock
func (t *TaskUserDataList) getUserDataList(userId string) *TasksUserData {
	d, exist := t.usersTaskData[userId]
	if !exist {
		//不存在则创建
		d = newTaskUserData(userId)
		t.usersTaskData[userId] = d
	}
	return d
}

func (t *TaskUserDataList) AddUserData(taskData *data.DashFunTaskUserData) {
	t.Lock()
	defer t.Unlock()
	l := t.getUserDataList(taskData.UserId)
	for idx, item := range l.taskDataList.Items() {
		if item.TaskId == taskData.TaskId {
			//当前记录中包含这个任务的进度数据，做更新
			l.taskDataList.RemoveAt(idx)
			l.taskDataList.Add(taskData)
			return
		}
	}
}
