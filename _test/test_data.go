package _test

import (
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/taskcenter"
	"log"
	"time"
)

func init() {
	makeTestTask()
	makeTestGame()
}

func makeTestTask() {
	t := taskcenter.Get()
	taskId := "8bzpjrva7ls"
	task := t.GetTaskById(taskId)
	if task == nil {
		//创建测试任务
		t.CreateTask(taskId, "Play \"War Three Kingdoms\"", "For Test")
	}
}

func makeTestGame() {
	s := gamecenter.Get()
	game, err := s.FindGame("6c2ghrcwm4g")
	if err != nil {
		log.Fatal(err)
	}

	if game == nil {
		//create game
		game = &data.DashFunGame{
			Id:   "6c2ghrcwm4g",
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

	game, err = s.FindGame("LocalTest")
	if game == nil {
		game = &data.DashFunGame{
			Id:   "LocalTest",
			Name: "LocalTest",
			Desc: "LocalTest",
			//Url:     "http://10.0.0.173:7456/web-mobile/web-mobile/index.html",
			Url:     "https://tma-game-test.nexgami.com/",
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
}
