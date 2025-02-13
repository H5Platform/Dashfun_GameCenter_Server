package data

type DashFunGame struct {
	Id          string            `json:"id" bson:"_id"`
	Name        string            `json:"name" bson:"name"`
	Desc        string            `json:"desc" bson:"desc"`
	Url         string            `json:"url" bson:"url"`                 //H5游戏部署地址
	Genre       []int             `json:"genre" bson:"genre"`             //游戏类型Id
	IconUrl     string            `json:"iconUrl" bson:"icon_url"`        //游戏图标地址
	LogoUrl     string            `json:"logoUrl" bson:"logo_url"`        //游戏logo
	MainPicUrl  string            `json:"mainPicUrl" bson:"main_pic_url"` //游戏主图地址 横向比例
	Time        int64             `json:"time" bson:"time"`               //游戏入库时间
	OpenTime    int64             `json:"openTime" bson:"open_time"`      //游戏开放时间，0表示开放，>0表示到达指定时间后才开放
	Status      DashFunGameStatus `json:"status" bson:"status"`           //游戏状态
	NewFlag     int               `json:"-" bson:"new_flag"`              //新游戏标志，会强出现在 New Game List中
	PopularFlag int               `json:"-" bson:"popular_flag"`          //最流行游戏标志，会强制出现在Popular List中
	SuggestFlag int               `json:"suggest" bson:"suggest_flag"`    //推荐游戏，出现在推荐列表中
	BannerFlag  int               `json:"-" bson:"banner_flag"`           //首屏游戏，出现在GameCenter顶部
	OpenCount   int               `json:"-" bson:"open_count"`            //游戏被开启的次数
	ApiSecret   string            `json:"-" bson:"api_secret"`
}

func (g *DashFunGame) IsTesting() bool {
	return g.Status <= DashFunGameStatus_Pending
}

type DashFunGameGenre struct {
	Id     int    `json:"id" bson:"_id"`
	Name   string `json:"name" bson:"name"`
	Hidden bool   `json:"hidden" bson:"hidden"` //true表示不显示在分类列表里，特殊分类，比如New, Popular等
}

type DashFunGameStatus int

const (
	DashFunGameStatus_NoChange DashFunGameStatus = iota //给update使用的，无变化
	DashFunGameStatus_Pending                           //刚创建的游戏，等待审核，pending状态的游戏充值不扣费
	DashFunGameStatus_Online                            //审核上线的游戏
	DashFunGameStatus_Removed                           //已经下线的游戏
)

type GameListType int

const (
	GameListType_Played   GameListType = iota //用户玩过的游戏列表，储存在OnlineUser中
	GameListType_New                          //最新游戏列表，按入库时间倒序
	GameListType_Popular                      //最流行游戏列表，按游玩次数倒序
	GameListType_Suggest                      //推荐游戏列表，suggest = 1
	GameListType_Banner                       //主屏顶部列表, banner=1
	GameListType_Favorite                     //用户收藏的游戏列表

	GameListTypeEnd = GameListType_Favorite
)

var Genres map[int]DashFunGameGenre

func init() {
	Genres = map[int]DashFunGameGenre{
		1: {
			Id:     1,
			Name:   "New",
			Hidden: true,
		},
		2: {
			Id:     2,
			Name:   "Popular",
			Hidden: true,
		},
		1001: {
			Id:     1001,
			Name:   "RPG",
			Hidden: false,
		},
		1002: {
			Id:     1002,
			Name:   "Card",
			Hidden: false,
		},
		1003: {
			Id:     1003,
			Name:   "Strategy",
			Hidden: false,
		},
	}
}
