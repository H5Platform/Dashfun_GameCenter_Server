package types

import (
	"dashfun_gamecenter/admin"
)

type AdminUserDao interface {
	FindUserById(id string) (*admin.AdminUser, error)
	FindUserByName(name string) (*admin.AdminUser, error)
	SaveUser(user *admin.AdminUser) (*admin.AdminUser, error)
}

type AdminUserLoginInfoDao interface {
	FindUserLoginInfo(id string) (*admin.AdminUserLoginInfo, error)
	SaveUserLoginInfo(info *admin.AdminUserLoginInfo) (*admin.AdminUserLoginInfo, error)
}
