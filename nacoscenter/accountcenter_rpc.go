package nacoscenter

import (
	"context"
	v1 "github.com/dashfun_web3/api_proto/gen/accountservice/v1"
	"go.uber.org/zap"
	"time"
)

type DashFunAccountStatus int

const (
	DashFunAccountStatusUnvalidated DashFunAccountStatus = iota //未验证
	DashFunAccountStatusNormal                                  //正常
	DashFunAccountStatusFrozen                                  //冻结
	DashFunAccountStatusDeleted                                 //删除
)

type DashFunAccountType int

const (
	DashFunAccountTypeEmail    DashFunAccountType = iota + 1 //邮箱
	DashFunAccountTypeTelegram                               //telegram
)

type AccountCenterRpc struct {
}

type AccountResult struct {
	AccountId   string               `json:"account_id"` //账号Id
	Username    string               `json:"username"`
	Type        DashFunAccountType   `json:"type"`   //账号类型，默认是email
	Status      DashFunAccountStatus `json:"status"` //账号状态
	Token       string               `json:"token"`  //登录token
	DisplayName string               `json:"display_name"`
}

func GetAccountCenterRpc() *AccountCenterRpc {
	_, err := Get().GetUserServiceClient()
	if err != nil {
		zap.S().Errorw("GetUserCenterRpc", "err", err)
		return nil
	}
	return &AccountCenterRpc{}
}

func (uc *AccountCenterRpc) CreateAccount(username, password string, accType DashFunAccountType) (*AccountResult, error) {
	accountServiceClient, err := Get().GetAccountServiceClient()
	if err != nil {
		zap.S().Errorw("GetAccountServiceClient", "err", err)
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &v1.CreateAccRequest{
		Username: username,
		Password: password,
		Type:     int32(accType),
	}

	resp, err := accountServiceClient.CreateAccount(ctx, req)
	if err != nil {
		zap.S().Errorw("CreateAccount", "username", username, "password", password, "type", accType, "err", err)
		return nil, err
	}

	return createAccPb2AccountResult(resp), nil
}

func (uc *AccountCenterRpc) LoginAccount(username, password string, accType DashFunAccountType) (*AccountResult, error) {
	accountServiceClient, err := Get().GetAccountServiceClient()
	if err != nil {
		zap.S().Errorw("GetAccountServiceClient", "err", err)
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &v1.AccountLoginRequest{
		Username: username,
		Password: password,
		Type:     int32(accType),
	}

	resp, err := accountServiceClient.AccountLogin(ctx, req)
	if err != nil {
		zap.S().Errorw("LoginAccount", "username", username, "password", password, "err", err)
		return nil, err
	}

	return loginAccPb2AccountResult(resp), nil
}

func (uc *AccountCenterRpc) RequestSendVerifyEmail(accountId string) error {
	accountServiceClient, err := Get().GetAccountServiceClient()
	if err != nil {
		zap.S().Errorw("GetAccountServiceClient", "err", err)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &v1.AccRequest{
		AccountId: accountId,
	}

	_, err = accountServiceClient.RequestSendVerifyEmail(ctx, req)

	if err != nil {
		zap.S().Errorw("RequestSendVerifyEmail", "accountId", accountId, "err", err)
		return err
	}
	return nil
}

func (uc *AccountCenterRpc) VerifyEmail(accountId, verifyCode string) (*AccountResult, error) {
	accountServiceClient, err := Get().GetAccountServiceClient()
	if err != nil {
		zap.S().Errorw("GetAccountServiceClient", "err", err)
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &v1.VerifyEmailRequest{
		AccountId:  accountId,
		VerifyCode: verifyCode,
	}

	resp, err := accountServiceClient.VerifyEmail(ctx, req)

	if err != nil {
		zap.S().Errorw("VerifyEmail", "accountId", accountId, "verifyCode", verifyCode, "err", err)
		return nil, err
	}

	result := loginAccPb2AccountResult(resp)
	return result, nil
}

func (uc *AccountCenterRpc) CheckToken(accountId, token string, accType DashFunAccountType) (*AccountResult, error) {
	accountServiceClient, err := Get().GetAccountServiceClient()
	if err != nil {
		zap.S().Errorw("GetAccountServiceClient", "err", err)
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &v1.CheckTokenRequest{
		AccountId: accountId,
		Token:     token,
		Type:      int32(accType),
	}

	resp, err := accountServiceClient.CheckToken(ctx, req)

	if err != nil {
		zap.S().Errorw("CheckToken", "accountId", accountId, "token", token, "err", err)
		return nil, err
	}

	result := loginAccPb2AccountResult(resp)
	return result, nil
}

func (uc *AccountCenterRpc) RequestResetPassword(username string) error {
	accountServiceClient, err := Get().GetAccountServiceClient()
	if err != nil {
		zap.S().Errorw("GetAccountServiceClient", "err", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &v1.ResetPasswordCheck{
		Username: username,
		Type:     int32(DashFunAccountTypeEmail),
	}

	_, err = accountServiceClient.RequestResetPassword(ctx, req)
	if err != nil {
		zap.S().Errorw("RequestResetPassword", "username", username, "err", err)
		return err
	}
	return nil
}

func (uc *AccountCenterRpc) ResetPassword(username, code, newPassword string) error {

	accountServiceClient, err := Get().GetAccountServiceClient()
	if err != nil {
		zap.S().Errorw("GetAccountServiceClient", "err", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &v1.ResetPasswordReset{
		Username: username,
		Code:     code,
		Password: newPassword,
		Type:     int32(DashFunAccountTypeEmail),
	}
	_, err = accountServiceClient.ResetPassword(ctx, req)
	if err != nil {
		zap.S().Errorw("ResetPassword", "username", username, "code", code, "newPassword", newPassword, "err", err)
		return err
	}
	return nil
}

func createAccPb2AccountResult(resp *v1.CreateAccResponse) *AccountResult {
	return &AccountResult{
		AccountId: resp.AccountId,
		Username:  resp.Username,
		Type:      DashFunAccountType(resp.Type),
		Status:    DashFunAccountStatusNormal,
	}
}

func loginAccPb2AccountResult(resp *v1.AccountLoginResponse) *AccountResult {
	return &AccountResult{
		AccountId:   resp.AccountId,
		Username:    resp.Username,
		Type:        DashFunAccountType(resp.Type),
		Status:      DashFunAccountStatus(resp.Status),
		Token:       resp.Token,
		DisplayName: resp.DisplayName,
	}
}
