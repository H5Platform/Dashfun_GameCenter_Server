package admin

type AdminUser struct {
	Id            string `json:"id" bson:"_id"`
	Name          string `json:"name" bson:"name"`
	Password      string `json:"password" bson:"password"`
	CreateAt      string `json:"create_at" bson:"create_at"`
	Authorization int    `json:"authorization" bson:"authorization"`
}

type AdminUserLoginInfo struct {
	Id       string `json:"id" bson:"_id"`
	Token    string `json:"token" bson:"token"`
	CreateAt string `json:"create_at" bson:"create_at"`
}
