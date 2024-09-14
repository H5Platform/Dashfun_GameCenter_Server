package admin

type AdminUserAuth int
type AdminUserStatus int

const (
	AdminAuth_Admin AdminUserAuth = -1
	AdminAuth_User                = 0x01
	AdminAuth_Game                = 0x02
	AdminAuth_Task                = 0x04
)

const (
	AdminStatus_Normal AdminUserStatus = iota + 1
	AdminStatus_ResetPassword
	AdminStatus_Ban
)

type AdminUser struct {
	Id            string          `json:"id" bson:"_id"`
	Name          string          `json:"name" bson:"name"`
	Email         string          `json:"email" bson:"email"`
	Password      string          `json:"password" bson:"password"`
	CreateAt      int64           `json:"create_at" bson:"create_at"`
	Status        AdminUserStatus `json:"status" bson:"status"`
	Authorization AdminUserAuth   `json:"authorization" bson:"authorization"`
}

type AdminUserLoginInfo struct {
	Id       string `json:"id" bson:"_id"`
	Token    string `json:"token" bson:"token"`
	CreateAt int64  `json:"create_at" bson:"create_at"`
}
