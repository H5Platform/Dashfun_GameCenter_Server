package types

import (
	"dashfun_gamecenter/admin"
)

type AdminUserDao interface {
	FindUserById(id string) (*admin.AdminUser, error)
	SaveUser(user *admin.AdminUser) (*admin.AdminUser, error)
}

type AdminUserLoginInfoDao interface {
	FindUserLoginInfo(id string) (*admin.AdminUserLoginInfo, error)
	SaveUserLoginInfo(*admin.AdminUserLoginInfo) error
}
