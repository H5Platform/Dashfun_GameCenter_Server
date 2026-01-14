package nolandev

import (
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/snowflake"
	"dashfun_gamecenter/usercenter"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var once sync.Once
var instance *NolanDevCenter

type NolanDevCenter struct {
	idGen *snowflake.Worker
}

func Get() *NolanDevCenter {
	once.Do(func() {
		instance = &NolanDevCenter{}
		instance.init()
	})
	return instance
}

func (f *NolanDevCenter) init() {
	f.idGen = snowflake.Must(snowflake.GetWorker(data.WorkerNolanPostId))
}

func (f *NolanDevCenter) NewPostId() string {
	id := f.idGen.NextId()
	return "fp" + strconv.FormatInt(id, 36)
}

func (f *NolanDevCenter) BotPost(id, name, post, location string, time int64) (*data.NolanPostData, error) {
	postId := f.NewPostId()
	d := dao.GetNolanDevPostDao()
	p, err := d.SaveOrUpdate(&data.NolanPostData{
		UserId:     id,
		PostId:     postId,
		PosterName: name,
		Content:    post,
		CreatedAt:  time,
		Location:   location,
	})
	return p, err
}

func (f *NolanDevCenter) GetPostPointReward(post *data.NolanPostData) int {
	point := 1000 //每次发帖奖励1000积分
	if post.Location != "" {
		point += 500 //如果有定位，额外奖励500积分
	}
	point += rand.Intn(99) //增加随机奖励，0-99之间的随机数
	return point
}

// Post creates a new post for the user and returns the post ID.
func (f *NolanDevCenter) Post(userId, post, location, fish string) (*data.NolanPostData, error) {
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
	d := dao.GetNolanDevPostDao()
	p, err := d.SaveOrUpdate(&data.NolanPostData{
		UserId:     userId,
		PostId:     postId,
		PosterName: ou.Nickname,
		Content:    post,
		CreatedAt:  time.Now().UnixMilli(),
		Location:   location,
	})

	//发帖成功，给用户积分
	if err == nil {
		c, existed := coincenter.Get().GetCoinByName("NolanDevPoint")
		if existed {
			point := f.GetPostPointReward(p)
			coincenter.Get().AddUserCoinAmount(userId, c.Id, int32(point), "nolandev post reward", postId)
		}
	}

	return p, err
}

func (f *NolanDevCenter) GetPosts(limit int) ([]*data.NolanPostData, error) {
	d := dao.GetNolanDevPostDao()
	if limit <= 0 {
		limit = 50
	}
	posts, err := d.GetPosts(limit)
	return posts, err
}

func (f *NolanDevCenter) GetUserDailyPostRemaining(userId string) int64 {
	d := dao.GetNolanDevPostDao()
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
