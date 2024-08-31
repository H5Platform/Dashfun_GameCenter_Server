package taskcenter

import (
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/snowflake"
	"log"
	"strconv"
	"sync"
	"time"
)

var once sync.Once
var instance *TaskCenter

type TaskCenter struct {
	idGen            *snowflake.Worker
	tasks            map[string]*data.DashFunTaskData
	taskUserDataList *TaskUserDataList
}

func Get() *TaskCenter {
	once.Do(func() {
		instance = &TaskCenter{}
		instance.init()
	})
	return instance
}

func (t *TaskCenter) init() {
	t.idGen = snowflake.Must(snowflake.GetWorker(data.WorkerTaskId))
	t.tasks = make(map[string]*data.DashFunTaskData)
	t.taskUserDataList = NewTaskUserDataList()

	//Load all tasks from db
	tasks := dao.GetTaskDao().FindAllTasks()
	for _, task := range tasks {
		t.tasks[task.Id] = task
	}

	events.UserEnterGameEvents.On(t.onUserEnterGameEvent)
	events.UserLoginEvents.On(t.onUserLogin)
}

func (t *TaskCenter) newTasId() string {
	id := t.idGen.NextId()
	return strconv.FormatInt(id, 36)
}

func (t *TaskCenter) GetTaskById(taskId string) *data.DashFunTaskData {
	task, ok := t.tasks[taskId]
	if !ok {
		task1, err := dao.GetTaskDao().FindTaskById(taskId)
		if err == nil {
			t.tasks[taskId] = task1
			task = task1
		}
	}
	return task
}

func (t *TaskCenter) CreateTaskAutoId(name, gameId string, taskType data.DashFunTaskType, category data.DashFunTaskCategory,
	condition data.DashFunTaskCondition, reward data.DashFunTaskReward) (*data.DashFunTaskData, error) {
	return t.CreateTask(t.newTasId(), name, gameId, taskType, category, condition, reward)
}

func (t *TaskCenter) CreateTask(taskId, name, gameId string, taskType data.DashFunTaskType, category data.DashFunTaskCategory,
	condition data.DashFunTaskCondition, reward data.DashFunTaskReward) (*data.DashFunTaskData, error) {
	task, err := dao.GetTaskDao().CreateTask(
		taskId, name, gameId, taskType, category, condition, reward)

	if err != nil {
		return nil, err
	}

	t.tasks[task.Id] = task

	return task, nil
}

func (t *TaskCenter) UpdateTaskOpenState(taskId string, open bool) {
	task := t.GetTaskById(taskId)
	if task != nil && task.Open != open {
		task.Open = open
		dao.GetTaskDao().SaveOrUpdate(task)
	}
}

func (t *TaskCenter) LoadAllTaskUserData(userId string) error {
	if t.taskUserDataList.HasRecord(userId) {
		//已经加载过了
		return nil
	}

	tasksData, err := dao.GetTaskUserDao().FindAllTaskUserData(userId)
	if err != nil {
		return err
	}

	for _, task := range tasksData {
		t.taskUserDataList.AddUserData(task)
	}
	return nil
}

// GetAllTaskUserData 获取用户所有任务的进度数据
func (t *TaskCenter) GetAllTaskUserData(userId string) ([]*data.DashFunTaskUserData, error) {
	tasksData, err := dao.GetTaskUserDao().FindAllTaskUserData(userId)
	if err != nil {
		return nil, err
	}
	return tasksData, nil
}

// GetTaskUserData 获取用户对应任务的进度记录，如果没有则新建，同时检测是否需要重置
func (t *TaskCenter) GetTaskUserData(userId, taskId string) (*data.DashFunTaskUserData, error) {
	userData, err := dao.GetTaskUserDao().FindTaskUserData(taskId, userId)
	if err != nil {
		return nil, err
	}
	if userData == nil {
		//create new userdata
		userData = &data.DashFunTaskUserData{
			UserId:   userId,
			TaskId:   taskId,
			Progress: 0,
			Time:     time.Now().UnixMilli(),
		}
		dao.GetTaskUserDao().SaveOrUpdate(userData)
	}

	task := t.GetTaskById(taskId)
	taskTime := time.UnixMilli(userData.Time)
	nowTime := time.Now()
	td := taskTime.YearDay()
	ty := taskTime.Year()

	nd := nowTime.YearDay()
	ny := nowTime.Year()

	reset := false

	if task != nil {
		switch task.Type {
		case data.TaskType_Daily:
			//不是同一天则重置进度
			reset = ty != ny || td != nd
			break

		case data.TaskType_2Days:
			reset = nd-td >= 2
			break
		}
	}

	if reset {
		userData = &data.DashFunTaskUserData{
			UserId:   userId,
			TaskId:   taskId,
			Progress: 0,
			Time:     time.Now().UnixMilli(),
		}
		dao.GetTaskUserDao().SaveOrUpdate(userData)
	}

	return userData, nil
}

func init() {
	log.Printf("new task id %s \n", Get().newTasId())
}
