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
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var once sync.Once
var instance *AdminUserMgr

type AdminUserAuthToken struct {
	UserId string `json:"user_id"`
	Token  string `json:"token"`
}

type adminUserSession struct {
	user        *admin.AdminUser
	loginInfo   *admin.AdminUserLoginInfo
	refreshTime int64
}

type AdminUserMgr struct {
	sessions map[string]*adminUserSession
	idGen    *snowflake.Worker
	sync.RWMutex
}

//type AdminSession struct {
//	User      *admin.AdminUser
//	LoginInfo *admin.AdminUserLoginInfo
//}

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
	mgr.sessions = make(map[string]*adminUserSession)
	mgr.idGen = snowflake.Must(snowflake.GetWorker(data.WorkerAdminUserId))

	adminCfg := config.GetConfig().AdminCfg

	usrAdmin, err := dao.GetAdminUserDao().FindUserByName("admin")
	if err == nil || errors.Is(err, mongo.ErrNoDocuments) {
		if usrAdmin == nil {
			//create user admin
			usrAdmin, err = mgr.createUser(adminCfg.Name, adminCfg.Password, " ", admin.AdminAuth_Admin)
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

func (mgr *AdminUserMgr) getAdminUserSession(userId string) (*adminUserSession, error) {
	mgr.Lock()
	defer mgr.Unlock()
	session, ok := mgr.sessions[userId]
	if !ok {
		info, err := dao.GetAdminUserLoginInfoDao().FindUserLoginInfo(userId)
		if err != nil {
			return nil, err
		}
		user, err := dao.GetAdminUserDao().FindUserById(userId)
		if err != nil {
			return nil, err
		}
		session = &adminUserSession{
			user:        user,
			loginInfo:   info,
			refreshTime: time.Now().UnixMilli(),
		}
		mgr.sessions[userId] = session
	}
	session.refreshTime = time.Now().UnixMilli()
	return session, nil
}

func (mgr *AdminUserMgr) GetAdminUser(userId string) (*admin.AdminUser, error) {
	session, err := mgr.getAdminUserSession(userId)
	if err != nil {
		return nil, err
	}
	if session != nil {
		return session.user, nil
	}

	au, err := dao.GetAdminUserDao().FindUserById(userId)
	if err != nil {
		return nil, err
	}

	return au, nil
}

func (mgr *AdminUserMgr) sendResetPasswordMail(user *admin.AdminUser, token string) {
	url := mgr.getActiveAccountUrl(user.Id, token)
	pinpoint.Get().SendEmail("Active your account", user.Email, "Please active your account with the link below\n"+url)
}

func (mgr *AdminUserMgr) getActiveAccountUrl(id, token string) string {
	activeUrl := config.GetConfig().Web.Url
	if !strings.HasSuffix(activeUrl, "/") {
		activeUrl += "/"
	}

	t := &AdminUserAuthToken{
		UserId: id,
		Token:  token,
	}

	marshal, err := json.Marshal(t)
	if err != nil {
		return ""
	}
	t1 := base64.StdEncoding.EncodeToString(marshal)
	activeUrl += "activate/" + t1
	return activeUrl
}

// ActiveUser 用户激活账户，设置密码
func (mgr *AdminUserMgr) ActiveUser(user *admin.AdminUser, password string) (*admin.AdminUser, error) {
	if len(password) == 0 {
		return nil, errors.New("password is empty")
	}

	user.Password = password
	dao.GetAdminUserDao().SaveUser(user)

	_, err := mgr.newLoginInfo(user.Id, mgr.newToken())
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateUser 创建一个用户，并发送邮件进行激活，设置密码
func (mgr *AdminUserMgr) CreateUser(name, email string, auth admin.AdminUserAuth) (*admin.AdminUser, error) {

	au, err := dao.GetAdminUserDao().FindUserByName(name)
	if err != nil {
		return nil, err
	}
	if au != nil {
		return nil, errors.New(fmt.Sprintf("user with name %s already exists", name))
	}

	au, err = dao.GetAdminUserDao().FindUserByMail(email)
	if err != nil {
		return nil, err
	}
	if au != nil {
		return nil, errors.New(fmt.Sprintf("user with email %s already exists", email))
	}

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
	if err != nil {
		return nil, err
	}
	_, err = mgr.newLoginInfo(id, token)
	if err != nil {
		return nil, err
	}
	mgr.sendResetPasswordMail(user, token)

	return user, err
}

func (mgr *AdminUserMgr) ResetUserPassword(user *admin.AdminUser) error {
	token := mgr.newToken()
	_, err := mgr.newLoginInfo(user.Id, token)
	if err != nil {
		return err
	}
	user.Status = admin.AdminStatus_ResetPassword
	_, err = dao.GetAdminUserDao().SaveUser(user)
	if err != nil {
		return err
	}
	mgr.sendResetPasswordMail(user, token)
	return nil
}

// CheckToken 检测用户的id和token是否正确，正确返回用户数据
func (mgr *AdminUserMgr) CheckToken(userId, token string) (*admin.AdminUser, bool) {
	session, err := mgr.getAdminUserSession(userId)
	if err != nil {
		zap.S().Errorw("check token failed", "userId", userId, "token", token, "err", err)
		return nil, false
	}
	if session == nil {
		zap.S().Errorw("check token failed", "userId", userId, "token", token, "err", "user login info not found")
		return nil, false
	}
	if session.loginInfo.Token != token {
		zap.S().Errorw("check token failed", "userId", userId, "token", token, "err", "invalid token", "expecting token", session.loginInfo.Token)
		return nil, false
	}

	return session.user, true
}

func (mgr *AdminUserMgr) Login(name, password string) (*admin.AdminUser, *admin.AdminUserLoginInfo, error) {
	user, err := dao.GetAdminUserDao().FindUserByName(name)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, errors.New("incorrect username or password")
	}
	if user.Status == admin.AdminStatus_ResetPassword || user.Status == admin.AdminStatus_Ban {
		return nil, nil, errors.New("incorrect user status")
	}
	if user.Password == password {
		token := mgr.newToken()
		loginInfo, err := mgr.newLoginInfo(user.Id, token)
		if err != nil {
			return nil, nil, err
		}

		dao.GetAdminUserLoginInfoDao().SaveUserLoginInfo(loginInfo)
		mgr.Lock()
		defer mgr.Unlock()
		mgr.sessions[user.Id] = &adminUserSession{
			user:        user,
			loginInfo:   loginInfo,
			refreshTime: time.Now().UnixMilli(),
		}

		return user, loginInfo, nil
	} else {
		return nil, nil, errors.New("incorrect username or password")
	}
}

func (mgr *AdminUserMgr) UpdateUserBaseInfo(userId, username, email string, auth admin.AdminUserAuth) (*admin.AdminUser, error) {
	user, err := mgr.GetAdminUser(userId)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	if username != "" {
		user.Name = username
	}
	if email != "" {
		user.Email = email
	}
	user.Authorization = auth
	saveUser, err := dao.GetAdminUserDao().SaveUser(user)

	if err != nil {
		return nil, err
	}

	return saveUser, nil
}

func (mgr *AdminUserMgr) UpdateUserStatus(userId string, status admin.AdminUserStatus) (*admin.AdminUser, error) {
	user, err := mgr.GetAdminUser(userId)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	if user.Status != status {
		user.Status = status
		saveUser, err := dao.GetAdminUserDao().SaveUser(user)
		if err != nil {
			return nil, err
		}
		return saveUser, nil
	}
	return user, nil
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

func (mgr *AdminUserMgr) GetUserList(name, email string, status admin.AdminUserStatus, size int64, page int64) ([]*admin.AdminUser, int, error) {
	r, totalPages, err := dao.GetAdminUserDao().SearchUser(name, email, status, size, page)
	return r, totalPages, err
}
