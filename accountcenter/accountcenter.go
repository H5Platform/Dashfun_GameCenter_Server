package accountcenter

import (
	"dashfun_gamecenter/nacoscenter"
	"sync"
)

type AccountCenter struct {
	accountCenterRpc *nacoscenter.AccountCenterRpc
}

var onceAccountCenterRpc sync.Once
var instAccountCenterRpc *AccountCenter

func Get() *AccountCenter {
	onceAccountCenterRpc.Do(func() {
		instAccountCenterRpc = &AccountCenter{}
		instAccountCenterRpc.init()
	})
	return instAccountCenterRpc
}

func (a *AccountCenter) init() {
	a.accountCenterRpc = nacoscenter.GetAccountCenterRpc()
}

func (a *AccountCenter) CreateAccount(username, password string, accType nacoscenter.DashFunAccountType) (*nacoscenter.AccountResult, error) {
	return a.accountCenterRpc.CreateAccount(username, password, accType)
}

func (a *AccountCenter) LoginAccount(username, password string, accType nacoscenter.DashFunAccountType) (*nacoscenter.AccountResult, error) {
	acc, err := a.accountCenterRpc.LoginAccount(username, password, accType)
	return acc, err
}

func (a *AccountCenter) RequestSendVerifyEmail(accountId string) error {
	return a.accountCenterRpc.RequestSendVerifyEmail(accountId)
}

func (a *AccountCenter) VerifyEmailCode(accountId, code string) (*nacoscenter.AccountResult, error) {
	return a.accountCenterRpc.VerifyEmail(accountId, code)
}

func (a *AccountCenter) CheckToken(accountId, token string, accType nacoscenter.DashFunAccountType) (*nacoscenter.AccountResult, error) {
	return a.accountCenterRpc.CheckToken(accountId, token, accType)
}

func (a *AccountCenter) RequestResetPassword(username string) error {
	return a.accountCenterRpc.RequestResetPassword(username)
}

func (a *AccountCenter) ResetPassword(username, code, password string) error {
	return a.accountCenterRpc.ResetPassword(username, code, password)
}
