package nolandata

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/nolandev"
	"dashfun_gamecenter/usercenter"
	"math/rand"
	"sync"
	"time"
)

type NolanDevBot struct {
	Data       *data.NolanBotData // Bot Data
	sync.Mutex                    // Mutex for thread safety
}

const maxScore = 50000

// RandomBot 随机生成一个bot
func RandomBot() *NolanDevBot {
	bot := &data.NolanBotData{}

	uid := usercenter.Get().RequestUserId()
	bot.Id = "fvb" + uid

	//① 随机bot的地区，CA,USA,MX,EU	2:2:1:1
	regions := []int{1000, 1000, 2000, 2000, 3000, 4000}
	region := regions[rand.Intn(len(regions))]
	// 从FishingRegions中筛选出与region匹配的项
	var matchedRegions []PostRegion
	for _, reg := range PostRegions {
		if reg.ID >= region && reg.ID <= region+999 {
			matchedRegions = append(matchedRegions, reg)
		}
	}
	selectedRegion := matchedRegions[rand.Intn(len(matchedRegions))]
	bot.RegionId = selectedRegion.ID

	//② 随机bot名字
	name := TemplateNames[rand.Intn(len(TemplateNames))]
	bot.Name = name

	//③ 随机当前分数
	score := float64((rand.Intn(maxScore/2)+(maxScore/2))*500/500) + float64(rand.Intn(1000))
	bot.Score = int64(score)
	bot.Rank = 0

	//④ 随机发帖间隔天数
	bot.MinPostIntervalDays = int64(rand.Intn(3) + 1)                       //1-3天
	bot.MaxPostIntervalDays = bot.MinPostIntervalDays + int64(rand.Intn(3)) //1-6天

	//⑤ 随机活跃时间

	if bot.MinPostIntervalDays == 1 && bot.MaxPostIntervalDays == 1 {
		// 将发帖间隔是1天的用户，作为初始化post的用户，这个bot会在建立之后发一个post
		bot.ActiveTime = time.Now().Truncate(24*time.Hour).UnixMilli() + int64(rand.Intn(24*60*60*1000)) // 随机时间点
	} else {
		bot.RandomNextActiveTime()
	}

	// 激活bot
	bot.Status = data.BotStatus_Active

	return &NolanDevBot{
		Data: bot,
	}
}

func (b *NolanDevBot) DoTodayBehaviour() {
	bot := b.Data
	region := GetPostRegionByID(bot.RegionId)
	post := RandomPostByRegion(region)

	//随机发帖是否带位置
	withLocation := rand.Intn(2) == 0
	location := ""
	if withLocation {
		location = region.Region + ", " + region.Country
	}
	postData, _ := nolandev.Get().BotPost(bot.Id, bot.Name, post.Content, location, bot.ActiveTime)

	if postData != nil {
		point := nolandev.Get().GetPostPointReward(postData)
		// 更新bot的分数
		bot.Score += int64(point)

	}

	bot.RandomNextActiveTime() // 更新下次活跃时间
}
