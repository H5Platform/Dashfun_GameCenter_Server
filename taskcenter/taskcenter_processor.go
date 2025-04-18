package taskcenter

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/tgbot"
	"dashfun_gamecenter/usercenter"
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
	if userData.Status == data.TaskStatus_InProgress || userData.Status == data.TaskStatus_ReturnInProgress {
		switch task.Condition.Type {
		case data.TaskCondition_FollowX:
			//follow x类，进入pending状态，同时给一个随机数，2-3，来模拟验证，让用户反复操作保证follow
			if userData.Status == data.TaskStatus_InProgress {
				//inprogress状态时新接任务
				checkData := &data.TaskSaveDataFollowX{
					RandomCount: rand.Intn(1) + 2,
					CheckCount:  0,
				}
				bytes, err := json.Marshal(checkData)
				if err != nil {
					return nil, err
				}
				userData.SaveData = string(bytes)
			}
			userData.Status = data.TaskStatus_Verify_Pending
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

func (t *TaskCenter) taskRecordEnterDashFun(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData) bool {
	ret := false
	if task.Condition.Type == data.TaskCondition_EnterDashFun && userData.Status == data.TaskStatus_InProgress {
		if isDashFunTask(task) && userData.Status == data.TaskStatus_InProgress {
			userData.Status = data.TaskStatus_Completed
			ret = true
		}
	}
	return ret
}

func (t *TaskCenter) taskRecordDailyLogin(userId string, task *data.DashFunTaskData, userData *data.DashFunTaskUserData) bool {
	ret := false

	if task.Condition.Type == data.TaskCondition_DailyLogin && userData.Status == data.TaskStatus_InProgress {
		save := data.TaskSaveDailyLogin{
			Days:     0,
			NextTime: time.Now().UnixMilli(),
		}

		if userData.SaveData != "" {
			err := json.Unmarshal([]byte(userData.SaveData), &save)
			if err != nil {
				zap.S().Errorw("unmarshal task daily login fail", err, err.Error(), "user", userId, "task", task.Id, "game", task.GameId, "savedata", userData.SaveData)
				ret = false
				return ret
			}
		}

		currentTime := time.Now().UnixMilli()
		if currentTime >= save.NextTime {
			if currentTime < save.NextTime+24*60*60*1000 {
				//在24小时内，连续登录
				save.Days++
			} else {
				//超过24小时，重新开始
				save.Days = 1
			}
			nextDay := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+1, 0, 0, 0, 0, time.Now().Location())
			save.NextTime = nextDay.UnixMilli()
			userData.Progress = save.Days

			marshal, err := json.Marshal(save)
			if err != nil {
				zap.S().Errorw("marshal task save data error", "err", err.Error(), "user", userId, "task", task.Id, "game", task.GameId, "savedata", userData.SaveData)
			} else {
				userData.SaveData = string(marshal)
			}

			ret = true
		}

		if save.Days >= task.Condition.Count {
			userData.Status = data.TaskStatus_Completed
			ret = true
		}
	}

	return ret
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
				//随机一个重复次数，如果没达到重试次数，任务状态变回inprogress，迫使用户再次点击follow
				if save.CheckCount >= save.RandomCount {
					userData.Status = data.TaskStatus_Completed
				} else {
					userData.Status = data.TaskStatus_ReturnInProgress
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
			channelId, err := usercenter.Get().GetDashFunUserChannelId(user.Id, data.DF_UserFrom_TG)
			if err != nil {
				zap.S().Errorw("get user channel id error", "user", user, "error", err)
				return false
			}
			tgUserId, err := strconv.ParseInt(channelId, 10, 64)
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

			zap.S().Infow("tgbot GetChatMember", "user", user, "channel", channelId, "member", member)

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

// taskRecordPlaySpecificGame 给玩家记录一次指定游戏的次数
func (t *TaskCenter) taskRecordPlaySpecificGame(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData, gameId string) bool {
	//玩指定游戏
	ret := false
	if task.Condition.Type == data.TaskCondition_PlaySpecificGame {
		if task.Condition.Condition == gameId && userData.Status == data.TaskStatus_InProgress {
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
		//l, err := strconv.Atoi(task.Condition.Condition)
		//if err != nil {
		//	zap.S().Errorw("task condition config error", "err", err, "task", task)
		//	return false
		//}
		l := task.Condition.Count
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
	if task.Condition.Type == data.TaskCondition_SpendDiamonds && userData.Status == data.TaskStatus_InProgress {
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

func (t *TaskCenter) taskRecordUserTGPayment(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData, payment *data.DashFunPaymentData, gameId string) bool {
	if task.Condition.Type == data.TaskCondition_SpendTGStar && userData.Status == data.TaskStatus_InProgress {
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

func (t *TaskCenter) taskRecordUserRecharge(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData, recharge *data.DashFunRechargeData) bool {
	if task.Condition.Type == data.TaskCondition_Recharge && userData.Status == data.TaskStatus_InProgress {
		ret := false
		if userData.Progress < task.Condition.Count {
			userData.Progress = userData.Progress + recharge.Diamond
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

func (t *TaskCenter) taskRecordUserLeaderboardChanged(task *data.DashFunTaskData, userData *data.DashFunTaskUserData, rank int64, score float64) bool {
	if task.Condition.Type == data.TaskCondition_LeaderboardRank && userData.Status == data.TaskStatus_InProgress {
		ret := false
		//排行榜排名大于要求排名，任务继续
		userData.Progress = int(rank)
		ret = true
		//排行榜排名复合排名要求
		if userData.Progress <= task.Condition.Count {
			userData.Status = data.TaskStatus_Completed
			ret = true
		}
		return ret
	}
	return false
}

func (t *TaskCenter) taskVerifyWalletAddress(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData) bool {
	if task.Condition.Type == data.TaskCondition_BindWallet && userData.Status == data.TaskStatus_InProgress {
		ret := false
		addr, ok := user.WalletAddress[task.Condition.Condition]
		if ok && len(addr) > 0 {
			userData.Status = data.TaskStatus_Completed
			ret = true
		}
		return ret
	}
	return false
}

func (t *TaskCenter) taskRecordInviteFriend(user *data.DashFunUser, task *data.DashFunTaskData, userData *data.DashFunTaskUserData) bool {
	if task.Condition.Type == data.TaskCondition_InviteFriends && userData.Status == data.TaskStatus_InProgress {
		ret := false
		if userData.Progress < task.Condition.Count {
			userData.Progress = userData.Progress + 1
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
