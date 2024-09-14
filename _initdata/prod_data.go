package _initdata

import (
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/taskcenter"
	"log"
	"time"
)

func init() {
	if config.IsProd() || config.IsDev() {
		makeProdGames()
		makeProdCoins()
		makeProdTasks()
	}

}

func makeProdCoins() {
	c := coincenter.Get()
	_, exist := c.GetCoinByName("DashFunCoin")
	if !exist {
		_, err := c.CreateCoin("", "DashFunCoin", "Dash", "DashFun Coin", true, 100, make(map[string]string))
		if err != nil {
			log.Fatal(err)
		}
	}

	c = coincenter.Get()
	_, exist = c.GetCoinByName("DashFunPoint")
	if !exist {
		_, err := c.CreateCoin("", "DashFunPoint", "Point", "DashFun Point", false, 0, make(map[string]string))
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

	if game == nil {
		//create game
		game = &data.DashFunGame{
			Id:   "",
			Name: "War Three Kingdoms",
			Desc: "Easy and fast Three Kingdomes Idle RPG",
			//Url:     "http://10.0.0.173:7456/web-mobile/web-mobile/index.html",
			Url:        "https://entry-tma.3kweb3.com/",
			MainPicUrl: "https://res.dashfun.games/pics/3kweb3-main.jpg",
			LogoUrl:    "https://res.dashfun.games/pics/3kweb3-logo.png",
			Genre:      []int{1, 1001},
			IconUrl:    "https://res.dashfun.games/icons/3kweb3-512.jpg",
			Time:       time.Now().UnixMilli(),
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
	task := t.GetTaskByName(taskName)

	game, err := gamecenter.Get().FindGameByName(gameName)
	if err != nil {
		log.Fatal(err)
	}

	if task == nil {
		//创建任务
		taskc, err := t.CreateTaskAutoId("Play \"War Three Kingdoms\"", game.Id, data.TaskType_Daily, data.TaskCategory_Daily,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_PlayGame,
				Count:     1,
				Condition: "",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_DashFunPoint,
				Amount:     1.5,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskName = "Play Any Game"
	task = t.GetTaskByName(taskName)
	if task == nil {
		//创建任务
		taskc, err := t.CreateTaskAutoId(taskName, "", data.TaskType_Daily, data.TaskCategory_Daily,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_PlayRandomGame,
				Count:     2,
				Condition: "",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_DashFunPoint,
				Amount:     10,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}

	taskName = "Join DashFun Group"
	task = nil
	task = t.GetTaskByName(taskName)
	chatId := "-1002176516558" //dashfun official group
	if task == nil {
		//创建测试任务
		//chatId := "-1002198592933"
		taskc, err := t.CreateTaskAutoId(taskName, "", data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_JoinTGChannel,
				Count:     1,
				Condition: chatId,
				Link:      "https://t.me/dashfun_official",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_DashFunPoint,
				Amount:     10,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	} else {
		task.Condition.Condition = chatId
		task.Condition.Link = "https://t.me/dashfun_official"
		dao.GetTaskDao().SaveOrUpdate(task)
	}

	taskName = "Follow DashFun On X"
	task = nil
	task = t.GetTaskByName(taskName)
	if task == nil {
		//创建任务
		taskc, err := t.CreateTaskAutoId("Follow DashFun On X", "", data.TaskType_Normal, data.TaskCategory_Challenges,
			data.DashFunTaskCondition{
				Type:      data.TaskCondition_FollowX,
				Count:     1,
				Condition: "",
				Link:      "https://x.com/dashfun_web3",
			},
			data.DashFunTaskReward{
				RewardType: data.TaskRewardType_DashFunPoint,
				Amount:     10,
			},
		)
		if err != nil {
			log.Fatalf("create task fail, err:%v", err)
		}
		task = taskc
	}
}
