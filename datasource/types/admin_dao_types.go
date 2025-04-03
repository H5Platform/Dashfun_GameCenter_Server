package types

import (
	"dashfun_gamecenter/admin"
)

type AdminUserDao interface {
	FindUserById(id string) (*admin.AdminUser, error)
	FindUserByName(name string) (*admin.AdminUser, error)
	FindUserByMail(email string) (*admin.AdminUser, error)
	SaveUser(user *admin.AdminUser) (*admin.AdminUser, error)
	SearchUser(name, email string, status admin.AdminUserStatus, size, page int64) (users []*admin.AdminUser, totalPages int, err error)
}

type AdminUserLoginInfoDao interface {
	FindUserLoginInfo(id string) (*admin.AdminUserLoginInfo, error)
	SaveUserLoginInfo(info *admin.AdminUserLoginInfo) (*admin.AdminUserLoginInfo, error)
}
