package data

type DashFunGame struct {
	Id        string `json:"id" bson:"_id"`
	Name      string `json:"name" bson:"name"`
	Desc      string `json:"desc" bson:"desc"`
	Url       string `json:"url" bson:"url"`          //H5游戏部署地址
	Genre     []int  `json:"genre" bson:"genre"`      //游戏类型Id
	IconUrl   string `json:"iconUrl" bson:"icon_url"` //游戏图标地址
	Time      int64  `json:"time" bson:"time"`        //游戏入库时间
	ApiSecret string `json:"-" bson:"api_secret"`
}

type DashFunGameGenre struct {
	Id     int    `json:"id" bson:"_id"`
	Name   string `json:"name" bson:"name"`
	Hidden bool   `json:"hidden" bson:"hidden"` //true表示不显示在分类列表里，特殊分类，比如New,Popular等
}

var Genres map[int]DashFunGameGenre

func init() {
	Genres = map[int]DashFunGameGenre{
		1: {
			Id:     1,
			Name:   "New",
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
	}
}
