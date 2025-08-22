package data

type FishingPostData struct {
	PostId     string `bson:"_id" json:"postId"`             // 帖子ID
	UserId     string `bson:"user_id" json:"userId"`         // 用户ID
	PosterName string `bson:"poster_name" json:"posterName"` // 帖子发布者名称
	Content    string `bson:"content" json:"content"`        // 帖子内容
	CreatedAt  int64  `bson:"created_at" json:"createdAt"`   // 帖子创建时间
	Location   string `bson:"location" json:"location"`      // 帖子位置
	FishCatch  string `bson:"fishCatch" json:"fishCatch"`    // 钓鱼种类
}
