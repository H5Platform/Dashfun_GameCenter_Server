package taskcenter

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/tgbot"
	"encoding/json"
	"errors"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
	"math/rand"
	"strconv"
	"time"
)

func (t *TaskCenter) UserClickedTask(user *data.DashFunUser, taskId string, gameId string) (*data.DashFunTaskUserData, error) {
	task := t.GetTaskById(taskId)
	if task == nil {
		return nil, errors.New("task not found")
	}
	userData, err := t.GetTaskUserData(user.Id, taskId)
	if err != nil {
		return nil, err
	}
	changed := false
	if userData.Status == data.TaskStatus_InProgress {
		switch task.Condition.Type {
		case data.TaskCondition_FollowX:
			//follow x类，进入pending状态，同时给一个随机数，2-4，来模拟验证，让用户反复操作保证follow
			userData.Status = data.TaskStatus_Verify_Pending
			checkData := &data.TaskSaveDataFollowX{
				RandomCount: rand.Intn(2) + 2,
				CheckCount:  0,
			}
			bytes, err := json.Marshal(checkData)
			if err != nil {
				return nil, err
			}
			userData.SaveData = string(bytes)
			changed = true
			break

		case data.TaskCondition_JoinTGChannel:
			//join tg channel,点击后变成待验证状态
			userData.Status = data.TaskStatus_Verify_Pending
			changed = true
			break
		}
	}

	if changed {
		t.saveTaskUserData(userData)
	}

	return userData, nil
}

func (t *TaskCenter) UserVerifyTask(user *data.DashFunUser, taskId string, gameId string) (*data.DashFunTaskUserData, error) {
	task := t.GetTaskById(taskId)
	if task == nil {
		return nil, errors.New("task not found")
	}
	userData, err := t.GetTaskUserData(user.Id, taskId)
	if err != nil {
		return nil, err
	}

	changed := false
	switch task.Condition.Type {
	case data.TaskCondition_FollowX:
		changed = t.taskVerifyFollowX(user, task, userData, gameId)
	case data.TaskCondition_JoinTGChannel:
		changed = t.taskVerifyTGChannel(user, task, userData, gameId)
	}

	if changed {
		t.saveTaskUserData(userData)
	}

	return userData, nil
}

func (t *TaskCenter) UserVerifyTGChannel(user *data.DashFunUser, taskId string, gameId string) (*data.DashFunTaskUserData, error) {
	task := t.GetTaskById(taskId)
	if task == nil {
		return nil, errors.New("task not found")
	}
	userData, err := t.GetTaskUserData(user.Id, taskId)
	if err != nil {
		return nil, err
	}
	changed := t.taskVerifyTGChannel(user, task, userData, gameId)
	if changed {
		t.saveTaskUserData(userData)
	}

	return userData, nil
}

func (t *TaskCenter) taskVerifyFollowX(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData, gameId string) bool {
	ret := false
	if task.Condition.Type == data.TaskCondition_FollowX {
		if (isDashFunTask(task) || task.GameId == gameId) && userData.Status == data.TaskStatus_Verify_Pending {
			save := data.TaskSaveDataFollowX{}
			err := json.Unmarshal([]byte(userData.SaveData), &save)
			if err != nil {
				zap.S().Errorw("unmarshal task save data error", "err", err.Error(), "user", user.Id, "task", task.Id, "game", task.GameId, "savedata", userData.SaveData)
				ret = false
			} else {
				save.CheckCount++
				if save.CheckCount >= save.RandomCount {
					userData.Status = data.TaskStatus_Completed
				}
				marshal, err := json.Marshal(save)
				if err != nil {
					zap.S().Errorw("marshal task save data error", "err", err.Error(), "user", user.Id, "task", task.Id, "game", task.GameId, "savedata", userData.SaveData)
				} else {
					userData.SaveData = string(marshal)
				}
				ret = true
			}

		}
	}
	return ret
}

// taskVerifyTGChannel 验证用户是否加入了tg channel
func (t *TaskCenter) taskVerifyTGChannel(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData, gameId string) bool {
	ret := false
	if task.Condition.Type == data.TaskCondition_JoinTGChannel {
		if (isDashFunTask(task) || task.GameId == gameId) && userData.Status == data.TaskStatus_Verify_Pending {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			tgUserId, err := strconv.ParseInt(user.ChannelId, 10, 64)
			if err != nil {
				zap.S().Errorw("user telegram id error", "user", user)
				return false
			}
			member, err := tgbot.Bot().GetChatMember(ctx, &bot.GetChatMemberParams{
				ChatID: task.Condition.Condition,
				UserID: tgUserId,
			})
			if err != nil {
				zap.S().Errorw("tgbot GetChatMember error", "user", user, "error", err)
				return false
			}

			if member.Type == models.ChatMemberTypeOwner || member.Type == models.ChatMemberTypeAdministrator || member.Type == models.ChatMemberTypeMember {
				userData.Progress = task.Condition.Count
				userData.Status = data.TaskStatus_Completed
				ret = true
			}
		}
	}
	return ret
}

// taskRecordPlayGame 给玩家记录一次指定游戏的次数
func (t *TaskCenter) taskRecordPlayGame(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData, gameId string) bool {
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

// taskRecordPlayRandomGame 给玩家记录一次游戏次数
func (t *TaskCenter) taskRecordPlayRandomGame(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData, gameId string) bool {
	//玩指定游戏
	if task.Condition.Type == data.TaskCondition_PlayRandomGame && userData.Status == data.TaskStatus_InProgress {
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
func (t *TaskCenter) taskRecordPlayerLevelUp(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData, gameId string, playerLevel int) bool {
	if task.Condition.Type == data.TaskCondition_LevelUp && userData.Status == data.TaskStatus_InProgress {
		l, err := strconv.Atoi(task.Condition.Condition)
		if err != nil {
			zap.S().Errorw("task condition config error", "err", err, "task", task)
			return false
		}

		if playerLevel >= l {
			//满足条件
			userData.Progress = playerLevel
			userData.Status = data.TaskStatus_Completed
			return true
		} else if playerLevel > userData.Progress {
			userData.Progress = playerLevel
			return true
		}
	}
	return false
}

func (t *TaskCenter) taskRecordUserPayment(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData, payment *data.DashFunPaymentData, gameId string) bool {
	if task.Condition.Type == data.TaskCondition_SpendTGStars && userData.Status == data.TaskStatus_InProgress {
		ret := false
		if userData.Progress < task.Condition.Count {
			userData.Progress = userData.Progress + payment.Price
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
