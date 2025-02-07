package _initdata

import (
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/spinwheelcenter"
	"dashfun_gamecenter/taskcenter"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

func init() {
	if config.IsDev() || config.IsTest() {
		makeTestUser()
		makeTestGame()
		makeTestTask()
		makeTestCoins()
		makeTestSpinWheel()
	}
}

func makeTestSpinWheel() {
	gameId := "LocalTest"

	spinWheelData, err := spinwheelcenter.Get().GetSpinWheelForGame(gameId)
	if err != nil {
		log.Fatalf("GetSpinWheelForGame err: %v", err)
	}
	if spinWheelData == nil {
		spinWheelData, err = spinwheelcenter.Get().CreateWheelForGame("SpinWheelTest", gameId, []data.SpinWheelReward{
			{
				RewardIndex: 0,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 5,
				Weight:      12,
			}, {
				RewardIndex: 1,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 10,
				Weight:      10,
			}, {
				RewardIndex: 2,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 5,
				Weight:      12,
			}, {
				RewardIndex: 3,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 10,
				Weight:      10,
			}, {
				RewardIndex: 4,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 5,
				Weight:      12,
			}, {
				RewardIndex: 5,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 10,
				Weight:      10,
			}, {
				RewardIndex: 6,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 5,
				Weight:      12,
			}, {
				RewardIndex: 7,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 20,
				Weight:      8,
			}, {
				RewardIndex: 8,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 5,
				Weight:      12,
			}, {
				RewardIndex: 9,
				RewardType:  data.SpinWheelReward_GamePoint,
				RewardValue: 50,
				Weight:      2,
			},
		})
		if err != nil {
			log.Fatalf("CreateWheelForGame err: %v", err)
		}
	}
}

func makeTestUser() {
	user, err := dao.GetUserDao().GetUserById("LocalTestUser")
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Fatalf("Get user err: %v", err)
	}

	if user == nil {
		user = &data.DashFunUser{
			Id:          "LocalTestUser",
			ChannelId:   fmt.Sprintf("TestTgUserId"),
			DisplayName: fmt.Sprintf("%s %s", "User", "Test"),
			UserName:    "TestUser",
			AvatarUrl:   "",
			From:        data.DF_UserFrom_TG,
			CreateData:  time.Now().UnixMilli(),
			LoginTime:   time.Now().UnixMilli(),
			LogoffTime:  0,
		}
		dao.GetUserDao().SaveOrUpdate(user)
	}
}

func makeTestCoins() {
	c := coincenter.Get()
	_, exist := c.GetCoinByName("DashFunCoin")
	if !exist {
		_, err := c.CreateCoin("", "DashFunCoin", "DFC", "DashFunCoin", "", true, 100, make(map[string]string))
		if err != nil {
			log.Fatal(err)
		}
	}

	_, exist = c.GetCoinByName("DashFunPoint")
	if !exist {
		_, err := c.CreateCoin("", "DashFunPoint", "DFP", "DashFun Point", "", false, 0, make(map[string]string))
		if err != nil {
			log.Fatal(err)
		}
	}

	_, exist = c.GetCoinByName("W3KPoint")
	if !exist {
		_, err := c.CreateCoin("", "W3KPoint", "W3KP", "W3K Point", "LocalTest", false, 0, make(map[string]string))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func makeTestTask() {
	t := taskcenter.Get()
	//taskId := "8bzpjrva7ls"
	taskId := "LocalTestTask"
	task := t.GetTaskById(taskId)
	if task == nil {
		//创建测试任务
		taskc, err := t.CreateTask(taskId, "Play \"LocalTest\"", "LocalTest", data.TaskType_Daily, data.TaskCategory_Daily,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_PlayGame,
				Count:     1,
				Condition: "",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     0.5,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskId = "LocalTestLevelUpTask3"
	task = t.GetTaskById(taskId)
	if task == nil {
		//创建测试任务
		taskc, err := t.CreateTask(taskId, "Level Up to Level 3", "LocalTest", data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_LevelUp,
				Count:     3,
				Condition: "3",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     1,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskId = "LocalTestLevelUpTask10"
	task = t.GetTaskById(taskId)
	if task == nil {
		//创建测试任务
		taskc, err := t.CreateTask(taskId, "Level Up to Level 10", "LocalTest", data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_LevelUp,
				Count:     10,
				Condition: "10",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     1,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskId = "PlayStoneAge"
	task = t.GetTaskById(taskId)
	if task == nil {
		//创建测试任务
		taskc, err := t.CreateTask(taskId, "Play \"Stone Age\" Test", "StoneAgeTest", data.TaskType_Daily, data.TaskCategory_Daily,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_PlayGame,
				Count:     1,
				Condition: "",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_GamePoint,
				Amount:     5,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskId = "PlayAnyGamesTest"
	task = t.GetTaskById(taskId)
	if task == nil {
		//创建测试任务
		taskc, err := t.CreateTask(taskId, "Play any game Test", "", data.TaskType_Daily, data.TaskCategory_Daily,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_PlayRandomGame,
				Count:     2,
				Condition: "",
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

	taskId = "LocalTestChannelTask"
	task = nil
	task = t.GetTaskById(taskId)
	if task == nil {
		//创建测试任务
		chatId := "-1002198592933"
		taskc, err := t.CreateTask(taskId, "Join DashFun channel Test", "LocalTest", data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_JoinTGChannel,
				Count:     1,
				Condition: chatId,
				Link:      "https://t.me/+h79TJSlUaO03ZDdh",
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

	taskId = "LocalTestTwitterTask"
	task = nil
	task = t.GetTaskById(taskId)
	if task == nil {
		//创建测试任务
		taskc, err := t.CreateTask(taskId, "Follow DashFun On X Test", "LocalTest", data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_FollowX,
				Count:     1,
				Condition: "",
				Link:      "https://x.com/nexgami",
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

	taskId = "LocalTestSpendStars"
	task = nil
	task = t.GetTaskById(taskId)
	if task == nil {
		//创建测试任务
		taskc, err := t.CreateTask(taskId, "Spend 5 stars", "LocalTest", data.TaskType_Daily, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_SpendTGStars,
				Count:     5,
				Condition: "",
				Link:      "",
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
	taskId = "LocalTestSpendStars500"
	task = nil
	task = t.GetTaskById(taskId)
	if task == nil {
		//创建测试任务
		taskc, err := t.CreateTask(taskId, "Spend 500 stars", "LocalTest", data.TaskType_Daily, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_SpendTGStars,
				Count:     500,
				Condition: "",
				Link:      "",
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

	taskId = "LocalTestSpendStars50000"
	task = nil
	task = t.GetTaskById(taskId)
	if task == nil {
		//创建测试任务
		taskc, err := t.CreateTask(taskId, "Spend 50000 stars", "LocalTest", data.TaskType_Daily, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_SpendTGStars,
				Count:     50000,
				Condition: "",
				Link:      "",
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

	taskId = "LocalTestBindWallet"
	task = nil
	task = t.GetTaskById(taskId)
	if task == nil {
		//创建测试任务
		taskc, err := t.CreateTask(taskId, "Connect your ton wallet", "", data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_BindWallet,
				Count:     1,
				Condition: "Ton", //需要绑定的网络
				Link:      "",
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

func makeTestGame() {
	s := gamecenter.Get()
	game, err := s.FindGameByName("Stone Age")
	if err != nil {
		log.Fatal(err)
	}

	if game == nil {
		//create game
		game = &data.DashFunGame{
			Id:   "StoneAgeTest",
			Name: "Stone Age",
			Desc: "Stone Age game description....",
			//Url:     "http://10.0.0.173:7456/web-mobile/web-mobile/index.html",
			Url:     "https://stone-res.83you.com/StoneAge_20230413/game.html",
			Genre:   []int{1, 1001},
			IconUrl: "",
			Time:    time.Now().UnixMilli(),
		}
		g, err := dao.GetGameDao().SaveOrUpdate(game)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Game Saved : %v", g)
	}

	game, err = s.FindGameByName("LocalTest")

	ot, _ := time.ParseInLocation(time.DateTime, "2024-11-05 00:00:00", time.Local)
	if game == nil {
		game = &data.DashFunGame{
			Id:   "LocalTest",
			Name: "LocalTest",
			Desc: "LocalTest",
			//Url:     "http://10.0.0.173:7456/web-mobile/web-mobile/index.html",
			Url:        "https://tma-game-test.nexgami.com/",
			MainPicUrl: "https://res.dashfun.games/pics/3kweb3-main.jpg",
			LogoUrl:    "https://res.dashfun.games/pics/3kweb3-logo.png",
			Genre:      []int{1, 1001},
			IconUrl:    "https://res.dashfun.games/icons/3kweb3-512.jpg",
			Time:       time.Now().UnixMilli(),
			OpenTime:   ot.UnixMilli(),
		}
		g, err := dao.GetGameDao().SaveOrUpdate(game)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Game Saved : %v", g)
	}

	game, err = s.FindGameByName("Brain(H5 Test)")
	if err != nil {
		log.Fatal(err)
	}

	if game == nil {
		//create game
		game = &data.DashFunGame{
			Id:   "h5-test-brain",
			Name: "Brain(H5 Test)",
			Desc: "Only for test",
			//Url:     "http://10.0.0.173:7456/web-mobile/web-mobile/index.html",
			Url:        "https://h5-test-brain.dashfun.games",
			MainPicUrl: "",
			LogoUrl:    "",
			Genre:      []int{1, 1001},
			IconUrl:    "",
			Time:       time.Now().UnixMilli(),
			OpenTime:   0,
		}
		g, err := s.SaveGame(game)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Game Saved : %v", g)
	}

	game, err = s.FindGameByName("ForTest")
	if err != nil {
		log.Fatal(err)
	}

	if game == nil {
		//create game
		game = &data.DashFunGame{
			Id:         "ForTest",
			Name:       "ForTest",
			Desc:       "Only for test",
			Url:        "",
			MainPicUrl: "",
			LogoUrl:    "",
			Genre:      []int{1, 1001},
			IconUrl:    "",
			Time:       time.Now().UnixMilli(),
			OpenTime:   0,
		}
		g, err := s.SaveGame(game)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Game Saved : %v", g)
	}
}
