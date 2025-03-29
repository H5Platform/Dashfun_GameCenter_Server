package taskcenter

import (
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/snowflake"
	"errors"
	"go.uber.org/zap"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

var once sync.Once
var instance *TaskCenter

type TaskCenter struct {
	idGen            *snowflake.Worker
	tasks            map[string]*data.DashFunTaskData
	taskUserDataList *TaskUserDataList
	tasksLock        sync.RWMutex
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
	t.taskUserDataList = newTaskUserDataList()

	//Load all tasks from db
	tasks := dao.GetTaskDao().FindAllTasks()
	for _, task := range tasks {
		t.tasks[task.Id] = task
		//process old task data
		if task.Rewards == nil {
			task.Rewards = append(task.Rewards, task.Reward)
		}
	}

	events.UserEnterGameEvents.On(t.onUserEnterGameEvent)
	events.UserLoginEvents.On(t.onUserLogin)
	events.PlayerLevelUpEvents.On(t.onGameReportPlayerLevelUp)
	events.UserPaymentEvents.On(t.onUserPayment)
	events.UserBindAddressEvents.On(t.onUserBindAddress)
}

func (t *TaskCenter) newTasId() string {
	id := t.idGen.NextId()
	return strconv.FormatInt(id, 36)
}

func (t *TaskCenter) GetTaskById(taskId string) *data.DashFunTaskData {
	t.tasksLock.RLock()
	task, ok := t.tasks[taskId]
	t.tasksLock.RUnlock()
	if !ok {
		task1, err := dao.GetTaskDao().FindTaskById(taskId)
		if err == nil {
			t.tasksLock.Lock()
			defer t.tasksLock.Unlock()
			t.tasks[taskId] = task1
			task = task1
		}
	}
	return task
}

// GetTaskByName 直接读取数据库，慎用
func (t *TaskCenter) GetTaskByName(taskName string) *data.DashFunTaskData {
	task, err := dao.GetTaskDao().FindTaskByName(taskName)
	if err != nil {
		return nil
	}
	return task
}

func (t *TaskCenter) CreateTaskAutoId(name, gameId string, taskType data.DashFunTaskType, category data.DashFunTaskCategory,
	condition data.DashFunTaskCondition, rewards ...data.DashFunTaskReward) (*data.DashFunTaskData, error) {
	return t.CreateTask(t.newTasId(), name, gameId, taskType, category, condition, rewards...)
}

func (t *TaskCenter) CreateTask(taskId, name, gameId string, taskType data.DashFunTaskType, category data.DashFunTaskCategory,
	condition data.DashFunTaskCondition, rewards ...data.DashFunTaskReward) (*data.DashFunTaskData, error) {
	task, err := dao.GetTaskDao().CreateTask(
		taskId, name, gameId, taskType, category, condition, rewards)

	if err != nil {
		return nil, err
	}

	t.tasksLock.Lock()
	defer t.tasksLock.Unlock()
	t.tasks[task.Id] = task

	return task, nil
}

func (t *TaskCenter) UpdateTask(taskId string, name string, taskType data.DashFunTaskType, category data.DashFunTaskCategory,
	condition data.DashFunTaskCondition, reward []data.DashFunTaskReward, isOpen bool) (*data.DashFunTaskData, error) {
	if taskId == "" {
		return nil, errors.New("task Id is empty")
	}
	task := t.GetTaskById(taskId)
	if task == nil {
		return nil, errors.New("task not found")
	}
	if name != "" {
		task.Name = name
	}
	task.Type = taskType
	task.Category = category
	task.Condition = condition
	task.Rewards = reward
	task.Open = isOpen
	dao.GetTaskDao().SaveOrUpdate(task)
	return task, nil
}

func (t *TaskCenter) UpdateTaskOpenState(taskId string, open bool) {
	task := t.GetTaskById(taskId)
	if task != nil && task.Open != open {
		task.Open = open
		dao.GetTaskDao().SaveOrUpdate(task)
	}
}

func (t *TaskCenter) loadAllTaskUserData(userId string) (*TasksUserData, error) {
	d, ok := t.taskUserDataList.HasRecord(userId)
	if ok {
		//已经加载过了
		return d, nil
	}

	tasksData, err := dao.GetTaskUserDao().FindAllTaskUserData(userId)
	if err != nil {
		return nil, err
	}

	tud := t.taskUserDataList.GetTasksUserData(userId)
	for _, td := range tasksData {
		tud.AddOrUpdateUserData(td)
	}
	return tud, nil
}

// loadTaskUserData 获取用户对应任务的进度数据，没有进度返回nil
func (t *TaskCenter) loadTaskUserData(userId, taskId string) (*data.DashFunTaskUserData, error) {
	d, err := t.loadAllTaskUserData(userId)
	if err != nil {
		return nil, err
	}
	taskUserData := d.GetTaskUserData(taskId)
	return taskUserData, nil
}

func (t *TaskCenter) saveTaskUserData(data *data.DashFunTaskUserData) {
	tud := t.taskUserDataList.GetTasksUserData(data.UserId)
	data.Time = time.Now().UnixMilli()
	tud.AddOrUpdateUserData(data)
	dao.GetTaskUserDao().SaveOrUpdate(data)
}

// GetTaskUserData 获取用户对应任务的进度记录，如果没有则新建，同时检测是否需要重置
// 同时针对加入TG Channel的任务，会在获取进度时进行验证是否完成
func (t *TaskCenter) GetTaskUserData(userId, taskId string) (*data.DashFunTaskUserData, error) {
	userData, err := t.loadTaskUserData(userId, taskId)
	if err != nil {
		return nil, err
	}
	if userData == nil {
		//create new userdata
		userData = newTaskUserData(userId, taskId)
		t.saveTaskUserData(userData)
	}

	task := t.GetTaskById(taskId)
	taskTime := time.UnixMilli(userData.Time)
	taskTime = time.Date(taskTime.Year(), taskTime.Month(), taskTime.Day(), 0, 0, 0, 0, taskTime.Location())
	nowTime := time.Now()
	nowTime = time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, nowTime.Location())

	//td := taskTime.YearDay()
	//ty := taskTime.Year()
	//
	//nd := nowTime.YearDay()
	//ny := nowTime.Year()

	diff := nowTime.Sub(taskTime)
	diffDays := diff.Hours() / 24

	reset := false

	if task != nil {
		switch task.Type {
		case data.TaskType_Daily:
			reset = diffDays >= 1
		case data.TaskType_2Days:
			reset = diffDays >= 2
		case data.TaskType_3Days:
			reset = diffDays >= 3
		case data.TaskType_7Days:
			reset = diffDays >= 7
		}
	}

	if reset {
		userData = newTaskUserData(userId, taskId)
		t.saveTaskUserData(userData)
	}

	return userData, nil
}

// 检查用户的任务进度，对某些任务进行验证，
// 如果修改了用户的数据，返回true，否则返回false
func (t *TaskCenter) checkTaskProgress(task *data.DashFunTaskData, user *data.DashFunUser, userData *data.DashFunTaskUserData, gameId string) bool {
	changed := false
	//2024-11-07 取消了自动验证，改为手动验证了
	//if task != nil && task.Condition.Type == data.TaskCondition_JoinTGChannel && userData.Status == data.TaskStatus_Verify_Pending {
	//	//加入tg channel的任务，如果任务状态为verify_pending，则在获取列表时进行验证
	//	changed = t.taskVerifyTGChannel(user, task, userData, gameId)
	//
	//}

	if task != nil && task.Condition.Type == data.TaskCondition_BindWallet {
		changed = t.taskVerifyWalletAddress(user, task, userData)
	}
	return changed
}

// UserClaimReward 用户请求获取任务奖励
func (t *TaskCenter) UserClaimReward(user *data.DashFunUser, taskId string) (*data.DashFunTaskUserData, error) {
	task := t.GetTaskById(taskId)
	if task == nil {
		zap.S().Errorw("User Claim Reward Error", "user", user.Id, "taskId", taskId, "error", "task not found")
		return nil, errors.New("task not found")
	}

	userData, err := t.GetTaskUserData(user.Id, taskId)
	if err != nil {
		zap.S().Errorw("User Claim Reward Error", "user", user.Id, "taskId", taskId, "error", err)
		return nil, err
	}

	if userData.Status == data.TaskStatus_Completed {
		//可领取奖励
		t.addTaskRewards(task, userData)
		return userData, nil
	} else {
		err = errors.New("task status error")
		zap.S().Errorw("User Claim Reward Error", "user", user.Id, "task", task, "error", err)
		return nil, err
	}
}

func (t *TaskCenter) addTaskReward(taskId, gameId string, reward *data.DashFunTaskReward, userData *data.DashFunTaskUserData) bool {
	c := coincenter.Get()
	var coin *data.CoinData
	var exist bool

	if reward.RewardType == data.TaskRewardType_GamePoint {
		//查找游戏对应的coin信息
		coin, exist = c.GetCoinByGame(gameId)
	} else {
		coin, exist = c.GetCoinByName(data.TaskRewardType2CoinName(reward.RewardType))
		if !exist {
			zap.S().Errorw("task reward type coin not found", "task", taskId)
			return false
		}
	}

	_, err := coincenter.Get().AddUserCoinAmount(userData.UserId, coin.Id, reward.Amount)
	if err != nil {
		zap.S().Errorw("AddUserCoinAmount err", "task", taskId, "error", err)
		return false
	}

	zap.S().Infow("User Claimed Task Reward", "task", taskId, "reward", reward.Amount, "coin", coin.Name)
	return true
}

func (t *TaskCenter) addTaskRewards(task *data.DashFunTaskData, userData *data.DashFunTaskUserData) {
	//add reward
	for _, reward := range task.Rewards {
		if !t.addTaskReward(task.Id, task.GameId, &reward, userData) {
			return
		}
	}

	//change user data
	userData.Status = data.TaskStatus_Claimed
	t.saveTaskUserData(userData)
}

// GetUserTaskInfo 获取用户的任务信息
// 返回用户在对应游戏中可用的任务列表，以及任务对应的进度
// gameId == "all" 时返回所有任务数据
// gameId == "" | "-1" | "DashFun" 返回DashFun的公共任务
func (t *TaskCenter) GetUserTaskInfo(user *data.DashFunUser, gameId string) *data.UserTaskInfo {
	userId := user.Id
	tasks := make([]*data.DashFunTaskData, 0)             //可用任务列表
	dataMap := make(map[string]*data.DashFunTaskUserData) //用户任务数据

	t.tasksLock.RLock()
	defer t.tasksLock.RUnlock()

	for _, task := range t.tasks {
		// 250320修改，游戏的任务列表中不在下发DashFun的任务
		if task.Open && ((isDashFunTask(task) && (gameId == "" || gameId == "-1" || strings.EqualFold(gameId, "dashfun"))) || task.GameId == gameId || gameId == "all") {
			userData, err := t.GetTaskUserData(userId, task.Id)
			if err != nil {
				zap.S().Errorw("get user task data error", "user", userId, "task", task)
				continue
			}
			tasks = append(tasks, task)
			if t.checkTaskProgress(task, user, userData, gameId) {
				//用户数据被修改
				t.saveTaskUserData(userData)
			}
			dataMap[task.Id] = userData
		}
	}

	slices.SortFunc(tasks, func(a, b *data.DashFunTaskData) int {
		if a.Category != b.Category {
			return int(a.Category - b.Category)
		} else {
			saveA := dataMap[a.Id]
			saveB := dataMap[b.Id]
			if saveA == nil || saveB == nil {
				return int(a.CreateTime - b.CreateTime)
			} else {
				if (saveA.Status != saveB.Status) && (saveA.Status == data.TaskStatus_Claimed || saveB.Status == data.TaskStatus_Claimed) {
					return int(saveA.Status - saveB.Status)
				} else {
					return int(a.CreateTime - b.CreateTime)
				}
			}
		}
	})

	ret := &data.UserTaskInfo{
		Tasks:    tasks,
		UserData: dataMap,
	}
	return ret
}

// GetGameTasksBackend 获取游戏对应的任务，仅供后台使用
// gameId 为空串时返回所有针对DashFun的公共任务
func (t *TaskCenter) GetGameTasksBackend(gameId string) []*data.DashFunTaskData {
	t.tasksLock.RLock()
	defer t.tasksLock.RUnlock()

	ret := make([]*data.DashFunTaskData, 0)

	for _, task := range t.tasks {
		if gameId == "" || gameId == "-1" {
			if isDashFunTask(task) {
				ret = append(ret, task)
			}
		} else if gameId == task.GameId {
			ret = append(ret, task)
		}
	}

	slices.SortFunc(ret, func(a, b *data.DashFunTaskData) int {
		if a.Category != b.Category {
			return int(a.Category - b.Category)
		} else {
			return int(a.CreateTime - b.CreateTime)
		}
	})

	return ret
}

func (t *TaskCenter) SearchTaskBackend(gameId, keyword string, size, page int64) (tasks []*data.DashFunTaskData, totalPage int, err error) {
	return dao.GetTaskDao().SearchTask(gameId, keyword, size, page)
}

func isDashFunTask(task *data.DashFunTaskData) bool {
	if task.GameId == "" || task.GameId == "-1" {
		return true
	}
	return false
}

func init() {
}
