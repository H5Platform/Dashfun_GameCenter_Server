package fishingverse

import (
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/snowflake"
	"dashfun_gamecenter/usercenter"
	"github.com/pkg/errors"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var once sync.Once
var instance *FishingPostCenter

type FishingPostCenter struct {
	idGen *snowflake.Worker
}

func Get() *FishingPostCenter {
	once.Do(func() {
		instance = &FishingPostCenter{}
		instance.init()
	})
	return instance
}

func (f *FishingPostCenter) init() {
	f.idGen = snowflake.Must(snowflake.GetWorker(data.WorkerFishingPostId))
}

func (f *FishingPostCenter) NewPostId() string {
	id := f.idGen.NextId()
	return "fp" + strconv.FormatInt(id, 36)
}

func (f *FishingPostCenter) BotPost(id, name, post, location, fish string, time int64) (*data.FishingPostData, error) {
	postId := f.NewPostId()
	d := dao.GetFishingPostDao()
	p, err := d.SaveOrUpdate(&data.FishingPostData{
		UserId:     id,
		PostId:     postId,
		PosterName: name,
		Content:    post,
		CreatedAt:  time,
		Location:   location,
		FishCatch:  fish,
	})
	return p, err
}

func (f *FishingPostCenter) GetPostPointReward(post *data.FishingPostData) int {
	point := 1000 //每次发帖奖励1000积分
	if post.FishCatch != "" {
		point += 500 //如果有钓鱼种类，额外奖励500积分
	}
	point += rand.Intn(99) //增加随机奖励，0-99之间的随机数
	return point
}

// Post creates a new post for the user and returns the post ID.
func (f *FishingPostCenter) Post(userId, post, location, fish string) (*data.FishingPostData, error) {
	ou, err := usercenter.Get().GetDashFunUser(userId)
	if err != nil {
		return nil, err
	}

	if f.GetUserDailyPostRemaining(userId) > 0 {
		return nil, errors.New("daily post limit reached")
	}

	if ou.Nickname == "" {
		return nil, errors.New("user profile not set")
	}

	postId := f.NewPostId()
	d := dao.GetFishingPostDao()
	p, err := d.SaveOrUpdate(&data.FishingPostData{
		UserId:     userId,
		PostId:     postId,
		PosterName: ou.Nickname,
		Content:    post,
		CreatedAt:  time.Now().UnixMilli(),
		Location:   location,
		FishCatch:  fish,
	})

	//发帖成功，给用户积分
	if err == nil {
		c, existed := coincenter.Get().GetCoinByName("FishingPoint")
		if existed {
			point := f.GetPostPointReward(p)
			coincenter.Get().AddUserCoinAmount(userId, c.Id, int32(point), "fishing post reward", postId)
		}
	}

	return p, err
}

func (f *FishingPostCenter) GetPosts(limit int) ([]*data.FishingPostData, error) {
	d := dao.GetFishingPostDao()
	if limit <= 0 {
		limit = 50
	}
	posts, err := d.GetPosts(limit)
	return posts, err
}

func (f *FishingPostCenter) GetUserDailyPostRemaining(userId string) int64 {
	d := dao.GetFishingPostDao()
	t, _ := d.GetUserLatestPostTime(userId)
	now := time.Now()
	last := time.UnixMilli(t)
	//不是同一天，或者已经超过24小时，重置为0
	if now.Day() != last.Day() || now.Sub(last) >= 24*time.Hour {
		return 0
	}
	//同一天，计算剩余时间
	return int64(time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).Sub(now).Seconds())
}
