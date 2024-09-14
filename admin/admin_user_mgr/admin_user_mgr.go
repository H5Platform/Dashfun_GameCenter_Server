package admin_user_mgr

import (
	"dashfun_gamecenter/admin"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/pinpoint"
	"dashfun_gamecenter/snowflake"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"log"
	"strconv"
	"sync"
	"time"
)

var once sync.Once
var instance *AdminUserMgr

type AdminUserAuthToken struct {
	UserId string `json:"user_id"`
	Token  string `json:"token"`
}

type AdminUserMgr struct {
	loginInfo map[string]*admin.AdminUserLoginInfo
	idGen     *snowflake.Worker
}

func Get() *AdminUserMgr {
	once.Do(func() {
		mgr := &AdminUserMgr{}
		mgr.init()
		instance = mgr
	})
	return instance
}

func (mgr *AdminUserMgr) newUserId() string {
	id := mgr.idGen.NextId()
	return strconv.FormatInt(id, 36)
}

func (mgr *AdminUserMgr) newToken() string {
	return uuid.NewString()
}

func (mgr *AdminUserMgr) init() {
	mgr.loginInfo = make(map[string]*admin.AdminUserLoginInfo)
	mgr.idGen = snowflake.Must(snowflake.GetWorker(data.WorkerAdminUserId))

	adminCfg := config.GetConfig().AdminCfg

	usrAdmin, err := dao.GetAdminUserDao().FindUserByName("admin")
	if err == nil || errors.Is(err, mongo.ErrNoDocuments) {
		if usrAdmin == nil {
			//create user admin
			usrAdmin, err = mgr.createUser(adminCfg.Name, adminCfg.Password, "", admin.AdminAuth_Admin)
			if err != nil {
				log.Fatalf("create admin user failed : %v", err.Error())
			}
		}
	}
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func (mgr *AdminUserMgr) createUser(name, password, email string, auth admin.AdminUserAuth) (*admin.AdminUser, error) {
	id := mgr.newUserId()
	adminUser := &admin.AdminUser{
		Id:            id,
		Name:          name,
		Email:         email,
		Password:      password,
		CreateAt:      time.Now().UnixMilli(),
		Status:        admin.AdminStatus_Normal,
		Authorization: auth,
	}
	user, err := dao.GetAdminUserDao().SaveUser(adminUser)
	return user, err
}

func (mgr *AdminUserMgr) newLoginInfo(userId, token string) (*admin.AdminUserLoginInfo, error) {
	loginInfo := &admin.AdminUserLoginInfo{
		Id:       userId,
		Token:    token,
		CreateAt: time.Now().UnixMilli(),
	}
	info, err := dao.GetAdminUserLoginInfoDao().SaveUserLoginInfo(loginInfo)
	return info, err
}

// CreateUser 创建一个用户，并发送邮件进行激活，设置密码
func (mgr *AdminUserMgr) CreateUser(name, email string, auth admin.AdminUserAuth) (*admin.AdminUser, error) {
	id := mgr.newUserId()
	token := mgr.newToken()
	adminUser := &admin.AdminUser{
		Id:            id,
		Name:          name,
		Email:         email,
		Password:      "",
		CreateAt:      time.Now().UnixMilli(),
		Status:        admin.AdminStatus_ResetPassword,
		Authorization: auth,
	}
	user, err := dao.GetAdminUserDao().SaveUser(adminUser)
	if err == nil {
		_, err = mgr.newLoginInfo(id, token)
	}

	pinpoint.Get().SendEmail("Active your account", email, "Please active your account with the link below\n"+token)

	return user, err
}

// CheckToken 检测用户的id和token是否正确，正确返回用户数据
func (mgr *AdminUserMgr) CheckToken(userId, token string) (*admin.AdminUser, bool) {
	user, err := dao.GetAdminUserDao().FindUserById(userId)
	if err != nil {
		zap.S().Errorw("check token failed", "userId", userId, "token", token, "err", err)
		return nil, false
	}
	loginInfo, err := dao.GetAdminUserLoginInfoDao().FindUserLoginInfo(userId)
	if err != nil {
		zap.S().Errorw("check token failed", "userId", userId, "token", token, "err", err)
		return nil, false
	}
	if loginInfo == nil {
		zap.S().Errorw("check token failed", "userId", userId, "token", token, "err", "user login info not found")
		return nil, false
	}
	if loginInfo.Token != token {
		zap.S().Errorw("check token failed", "userId", userId, "token", token, "err", "invalid token", "expecting token", loginInfo.Token)
		return nil, false
	}

	return user, true
}

func (mgr *AdminUserMgr) Login(name, password string) (*admin.AdminUser, *admin.AdminUserLoginInfo, error) {
	user, err := dao.GetAdminUserDao().FindUserByName(name)
	if err != nil {
		return nil, nil, err
	}
	if user.Password == password {
		token := mgr.newToken()
		loginInfo, err := mgr.newLoginInfo(user.Id, token)
		if err != nil {
			return nil, nil, err
		}
		dao.GetAdminUserLoginInfoDao().SaveUserLoginInfo(loginInfo)
		return user, loginInfo, nil
	} else {
		return nil, nil, errors.New("invalid password")
	}
}

func (mgr *AdminUserMgr) ParseToken(authString string) (*AdminUserAuthToken, error) {
	t := &AdminUserAuthToken{}

	decodeString, err := base64.StdEncoding.DecodeString(authString)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(decodeString, t)
	if err != nil {
		return nil, err
	}

	return t, nil
}
