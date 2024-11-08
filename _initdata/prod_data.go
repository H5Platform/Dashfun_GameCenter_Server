package _initdata

import (
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/spinwheelcenter"
	"dashfun_gamecenter/taskcenter"
	"log"
	"time"
)

func init() {
	if config.IsProd() {
		makeProdGames()
		makeProdCoins()
		makeProdTasks()
		makeProdSpinWheel()
	}

}

func makeProdSpinWheel() {
	gameName := "War Three Kingdoms"

	w3kt, err := gamecenter.Get().FindGameByName(gameName)
	if err != nil {
		log.Fatal(err)
	}

	gameId := w3kt.Id

	spinWheelData, err := spinwheelcenter.Get().GetSpinWheelForGame(gameId)
	if err != nil {
		log.Fatalf("GetSpinWheelForGame err: %v", err)
	}
	if spinWheelData == nil {
		spinWheelData, err = spinwheelcenter.Get().CreateWheelForGame("SpinWheelTest", gameId, []data.SpinWheelReward{
			{
				RewardIndex: 0,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 2,
				Weight:      5,
			}, {
				RewardIndex: 1,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 4,
				Weight:      10,
			}, {
				RewardIndex: 2,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 6,
				Weight:      20,
			}, {
				RewardIndex: 3,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 8,
				Weight:      20,
			}, {
				RewardIndex: 4,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 10,
				Weight:      20,
			}, {
				RewardIndex: 5,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 12,
				Weight:      20,
			}, {
				RewardIndex: 6,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 14,
				Weight:      20,
			}, {
				RewardIndex: 7,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 16,
				Weight:      20,
			}, {
				RewardIndex: 8,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 18,
				Weight:      20,
			}, {
				RewardIndex: 9,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 100,
				Weight:      1,
			},
		})
		if err != nil {
			log.Fatalf("CreateWheelForGame err: %v", err)
		}
	}
}

func makeProdCoins() {
	gameName := "War Three Kingdoms"
	w3kt, err := gamecenter.Get().FindGameByName(gameName)
	if err != nil {
		log.Fatal(err)
	}

	c := coincenter.Get()
	_, exist := c.GetCoinByName("DashFunCoin")
	if !exist {
		_, err := c.CreateCoin("", "DashFunCoin", "Dash", "DashFun Coin", "", true, 100, make(map[string]string))
		if err != nil {
			log.Fatal(err)
		}
	}

	_, exist = c.GetCoinByName("DashFunPoint")
	if !exist {
		_, err := c.CreateCoin("", "DashFunPoint", "Point", "DashFun Point", "", false, 0, make(map[string]string))
		if err != nil {
			log.Fatal(err)
		}
	}

	_, exist = c.GetCoinByName("W3KPoint")
	if !exist {
		_, err := c.CreateCoin("", "W3KPoint", "W3KPoint", "W3K Point", w3kt.Id, false, 0, make(map[string]string))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func makeProdGames() {
	s := gamecenter.Get()
	game, err := s.FindGameByName("War Three Kingdoms")
	if err != nil {
		log.Fatal(err)
	}

	ot, _ := time.ParseInLocation(time.DateTime, "2024-11-10 11:00:00", time.Local)

	if game == nil {
		//create game
		game = &data.DashFunGame{
			Id:   "9c4r4sdzb40", //三国固定用这个id
			Name: "War Three Kingdoms",
			Desc: "Easy and fast Three Kingdomes Idle RPG",
			//Url:     "http://10.0.0.173:7456/web-mobile/web-mobile/index.html",
			Url:        "https://entry-tma.3kweb3.com/",
			MainPicUrl: "https://res.dashfun.games/pics/3kweb3-main.jpg",
			LogoUrl:    "https://res.dashfun.games/pics/3kweb3-logo.png",
			Genre:      []int{1, 1001},
			IconUrl:    "https://res.dashfun.games/icons/3kweb3-512.jpg",
			Time:       time.Now().UnixMilli(),
			OpenTime:   ot.UnixMilli(),
		}
		g, err := s.SaveGame(game)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Game Saved : %v", g)
	}
}

func makeProdTasks() {
	t := taskcenter.Get()

	gameName := "War Three Kingdoms"
	taskName := "Play \"War Three Kingdoms\""

	w3kt, err := gamecenter.Get().FindGameByName(gameName)
	if err != nil {
		log.Fatal(err)
	}

	task := t.GetTaskByName(taskName)
	if task == nil {
		//创建任务
		taskc, err := t.CreateTaskAutoId("Play \"War Three Kingdoms\"", w3kt.Id, data.TaskType_Daily, data.TaskCategory_Daily,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_PlayGame,
				Count:     1,
				Condition: "",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     20,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	//1. follow twitter (can not verify because users can't login twitter in TMA)
	//2. join tg group
	//3. play w3k (can be a daily task)
	//4. level up to level (can create multiple tasks, eg. Level up to Lv.3   Level up to Lv.10)
	//5. spend tg stars (can be a daily task , eg. spend 100 stars a day)
	//

	taskName = "Follow Official Twitter"
	task = nil
	task = t.GetTaskByName(taskName)
	if task == nil {
		//创建任务
		taskc, err := t.CreateTaskAutoId(taskName, w3kt.Id, data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_FollowX,
				Count:     1,
				Condition: "",
				Link:      "https://x.com/war3kingdom",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     10,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskName = "Follow NexGami"
	task = nil
	task = t.GetTaskByName(taskName)
	if task == nil {
		//创建任务
		taskc, err := t.CreateTaskAutoId(taskName, w3kt.Id, data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_FollowX,
				Count:     1,
				Condition: "",
				Link:      "https://x.com/nexgami",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     8,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskName = "Join Official Group"
	task = nil
	task = t.GetTaskByName(taskName)
	chatId := "@war3kingdoms" //dashfun official group
	if task == nil {
		//创建测试任务
		//chatId := "-1002198592933"
		taskc, err := t.CreateTaskAutoId(taskName, w3kt.Id, data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_JoinTGChannel,
				Count:     1,
				Condition: chatId,
				Link:      "https://t.me/war3kingdoms",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     15,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskName = "Join NexGami Group"
	task = nil
	task = t.GetTaskByName(taskName)
	chatId = "@nexgami" //dashfun official group
	if task == nil {
		//创建测试任务
		//chatId := "-1002198592933"
		taskc, err := t.CreateTaskAutoId(taskName, w3kt.Id, data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_JoinTGChannel,
				Count:     1,
				Condition: chatId,
				Link:      "https://t.me/nexgami",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     12,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	makeLevelUpTasks(w3kt)
	makeSpendStarTasks(w3kt)
	//taskName = "Play Any Game"
	//task = t.GetTaskByName(taskName)
	//if task == nil {
	//	//创建任务
	//	taskc, err := t.CreateTaskAutoId(taskName, "", data.TaskType_Daily, data.TaskCategory_Daily,
	//		data.DashFunTaskCondition{
	//			Type:      data.TaskCondition_PlayRandomGame,
	//			Count:     2,
	//			Condition: "",
	//		},
	//		data.DashFunTaskReward{
	//			RewardType: data.TaskRewardType_GamePoint,
	//			Amount:     10,
	//		},
	//	)
	//	if err != nil {
	//		log.Fatalf("create task fail, err:%v", err)
	//	}
	//	task = taskc
	//}
	//
	//taskName = "Join DashFun Group"
	//task = nil
	//task = t.GetTaskByName(taskName)
	//chatId := "@dashfun_official" //dashfun official group
	//if task == nil {
	//	//创建测试任务
	//	//chatId := "-1002198592933"
	//	taskc, err := t.CreateTaskAutoId(taskName, "", data.TaskType_Normal, data.TaskCategory_Challenges,
	//		data.DashFunTaskCondition{
	//			Type:      data.TaskCondition_JoinTGChannel,
	//			Count:     1,
	//			Condition: chatId,
	//			Link:      "https://t.me/dashfun_official",
	//		},
	//		data.DashFunTaskReward{
	//			RewardType: data.TaskRewardType_GamePoint,
	//			Amount:     10,
	//		},
	//	)
	//	if err != nil {
	//		log.Fatalf("create task fail, err:%v", err)
	//	}
	//	task = taskc
	//} else {
	//	task.Condition.Condition = chatId
	//	task.Condition.Link = "https://t.me/dashfun_official"
	//	dao.GetTaskDao().SaveOrUpdate(task)
	//}
	//
	//taskName = "Follow DashFun On X"
	//task = nil
	//task = t.GetTaskByName(taskName)
	//if task == nil {
	//	//创建任务
	//	taskc, err := t.CreateTaskAutoId("Follow DashFun On X", "", data.TaskType_Normal, data.TaskCategory_Challenges,
	//		data.DashFunTaskCondition{
	//			Type:      data.TaskCondition_FollowX,
	//			Count:     1,
	//			Condition: "",
	//			Link:      "https://x.com/dashfun_web3",
	//		},
	//		data.DashFunTaskReward{
	//			RewardType: data.TaskRewardType_GamePoint,
	//			Amount:     10,
	//		},
	//	)
	//	if err != nil {
	//		log.Fatalf("create task fail, err:%v", err)
	//	}
	//	task = taskc
	//}
	//
	//taskName = "Spend 100 stars"
	//task = nil
	//task = t.GetTaskByName(taskName)
	//if task == nil {
	//	//创建测试任务
	//	taskc, err := t.CreateTaskAutoId(taskName, w3kt.Id, data.TaskType_Daily, data.TaskCategory_Daily,
	//		data.DashFunTaskCondition{
	//			Type:      data.TaskCondition_SpendTGStars,
	//			Count:     100,
	//			Condition: "",
	//			Link:      "",
	//		},
	//		data.DashFunTaskReward{
	//			RewardType: data.TaskRewardType_GamePoint,
	//			Amount:     10,
	//		},
	//	)
	//	if err != nil {
	//		log.Fatalf("create task fail, err:%v", err)
	//	}
	//	task = taskc
	//}
	//
	//taskName = "Spend 1000 stars"
	//task = nil
	//task = t.GetTaskByName(taskName)
	//if task == nil {
	//	//创建测试任务
	//	taskc, err := t.CreateTaskAutoId(taskName, w3kt.Id, data.TaskType_Daily, data.TaskCategory_Daily,
	//		data.DashFunTaskCondition{
	//			Type:      data.TaskCondition_SpendTGStars,
	//			Count:     1000,
	//			Condition: "",
	//			Link:      "",
	//		},
	//		data.DashFunTaskReward{
	//			RewardType: data.TaskRewardType_GamePoint,
	//			Amount:     100,
	//		},
	//	)
	//	if err != nil {
	//		log.Fatalf("create task fail, err:%v", err)
	//	}
	//	task = taskc
	//}
	//
	//taskName = "Level Up to Lv.10"
	//task = t.GetTaskByName(taskName)
	//if task == nil {
	//	//创建任务
	//	taskc, err := t.CreateTaskAutoId(taskName, w3kt.Id, data.TaskType_Normal, data.TaskCategory_Challenges,
	//		data.DashFunTaskCondition{
	//			Type:      data.TaskCondition_LevelUp,
	//			Count:     10,
	//			Condition: "10",
	//		},
	//		data.DashFunTaskReward{
	//			RewardType: data.TaskRewardType_GamePoint,
	//			Amount:     100,
	//		},
	//	)
	//	if err != nil {
	//		log.Fatalf("create task fail, err:%v", err)
	//	}
	//	task = taskc
	//}
}
func makeSpendStarTasks(game *data.DashFunGame) {
	t := taskcenter.Get()

	taskName := "Spend 100 TG Stars"
	task := t.GetTaskByName(taskName)
	if task == nil {
		//创建测试任务
		taskc, err := t.CreateTaskAutoId(taskName, game.Id, data.TaskType_Daily, data.TaskCategory_Daily,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_SpendTGStars,
				Count:     100,
				Condition: "",
				Link:      "",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     30,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskName = "Spend 200 TG Stars"
	task = t.GetTaskByName(taskName)
	if task == nil {
		//创建测试任务
		taskc, err := t.CreateTaskAutoId(taskName, game.Id, data.TaskType_Daily, data.TaskCategory_Daily,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_SpendTGStars,
				Count:     200,
				Condition: "",
				Link:      "",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     50,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskName = "Spend 1000 TG Stars over 7 days"
	task = t.GetTaskByName(taskName)
	if task == nil {
		//创建测试任务
		taskc, err := t.CreateTaskAutoId(taskName, game.Id, data.TaskType_Daily, data.TaskCategory_Daily,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_SpendTGStars,
				Count:     1000,
				Condition: "",
				Link:      "",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     150,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}
}

func makeLevelUpTasks(game *data.DashFunGame) {
	t := taskcenter.Get()

	taskName := "Level Up to Lv.3"
	task := t.GetTaskByName(taskName)
	if task == nil {
		//创建任务
		taskc, err := t.CreateTaskAutoId(taskName, game.Id, data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_LevelUp,
				Count:     3,
				Condition: "3",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     25,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskName = "Level Up to Lv.5"
	task = t.GetTaskByName(taskName)
	if task == nil {
		//创建任务
		taskc, err := t.CreateTaskAutoId(taskName, game.Id, data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_LevelUp,
				Count:     5,
				Condition: "5",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     40,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskName = "Level Up to Lv.10"
	task = t.GetTaskByName(taskName)
	if task == nil {
		//创建任务
		taskc, err := t.CreateTaskAutoId(taskName, game.Id, data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_LevelUp,
				Count:     10,
				Condition: "10",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     60,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskName = "Level Up to Lv.20"
	task = t.GetTaskByName(taskName)
	if task == nil {
		//创建任务
		taskc, err := t.CreateTaskAutoId(taskName, game.Id, data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_LevelUp,
				Count:     20,
				Condition: "20",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     100,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

}
