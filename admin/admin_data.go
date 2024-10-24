package admin

type AdminUserAuth int
type AdminUserStatus int

const (
	AdminAuth_None  AdminUserAuth = 0 //某些api不限制权限，只要求登录的时候，使用这个权限
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

type AdminUserResult struct {
	Id            string          `json:"id" bson:"_id"`
	Name          string          `json:"name" bson:"name"`
	Email         string          `json:"email" bson:"email"`
	CreateAt      int64           `json:"create_at" bson:"create_at"`
	Status        AdminUserStatus `json:"status" bson:"status"`
	Authorization AdminUserAuth   `json:"authorization" bson:"authorization"`
}

func ToAdminUserResult(user *AdminUser) *AdminUserResult {
	return &AdminUserResult{
		Id:            user.Id,
		Name:          user.Name,
		Email:         user.Email,
		CreateAt:      user.CreateAt,
		Status:        user.Status,
		Authorization: user.Authorization,
	}
}

type AdminUserLoginInfo struct {
	Id       string `json:"id" bson:"_id"`
	Token    string `json:"token" bson:"token"`
	CreateAt int64  `json:"create_at" bson:"create_at"`
}
