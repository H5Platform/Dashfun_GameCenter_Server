package taskcenter

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/utils"
	"sync"
	"time"
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

func newTaskUserDataList() *TaskUserDataList {
	return &TaskUserDataList{
		usersTaskData: make(map[string]*TasksUserData),
	}
}

func newTasksUserData(userId string) *TasksUserData {
	return &TasksUserData{
		userId:       userId,
		taskDataList: utils.NewList[*data.DashFunTaskUserData](),
	}
}

func newTaskUserData(userId, taskId string) *data.DashFunTaskUserData {
	return &data.DashFunTaskUserData{
		UserId:   userId,
		TaskId:   taskId,
		Progress: 0,
		SaveData: "",
		Status:   data.TaskStatus_InProgress,
		Time:     time.Now().UnixMilli(),
	}
}

func (t *TaskUserDataList) HasRecord(userId string) (*TasksUserData, bool) {
	t.RLock()
	defer t.RUnlock()
	d, ok := t.usersTaskData[userId]
	return d, ok
}

func (t *TaskUserDataList) GetTasksUserData(userId string) *TasksUserData {
	t.Lock()
	defer t.Unlock()
	d, exist := t.usersTaskData[userId]
	if !exist {
		//不存在则创建
		d = newTasksUserData(userId)
		t.usersTaskData[userId] = d
	}
	return d
}

func (t *TaskUserDataList) RemoveTasksUserData(userId string) {
	t.Lock()
	defer t.Unlock()
	_, exist := t.usersTaskData[userId]
	if exist {
		t.usersTaskData[userId] = nil
		delete(t.usersTaskData, userId)
	}
}

func (tud *TasksUserData) AddUserData(taskData *data.DashFunTaskUserData) {
	for idx, item := range tud.taskDataList.Items() {
		if item.TaskId == taskData.TaskId {
			//当前记录中包含这个任务的进度数据，做更新
			tud.taskDataList.RemoveAt(idx)
			tud.taskDataList.Add(taskData)
			return
		}
	}
	tud.taskDataList.Add(taskData)
}

func (tud *TasksUserData) GetTaskUserData(taskId string) *data.DashFunTaskUserData {
	for _, userData := range tud.taskDataList.Items() {
		if userData.TaskId == taskId {
			return userData
		}
	}
	return nil
}
