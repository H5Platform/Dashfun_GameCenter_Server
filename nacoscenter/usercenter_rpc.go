package nacoscenter

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/utils"
	"errors"
	"github.com/dashfun_web3/api_proto/gen/common"
	v1 "github.com/dashfun_web3/api_proto/gen/userservice/v1"
	"go.uber.org/zap"
	"time"
)

type UserCenterRpc struct {
}

func GetUserCenterRpc() *UserCenterRpc {
	_, err := Get().GetUserServiceClient()
	if err != nil {
		zap.S().Errorw("GetUserCenterRpc", "err", err)
		return nil
	}
	return &UserCenterRpc{}
}

func (u *UserCenterRpc) GetDashFunUser(userId string) (*data.DashFunUser, error) {
	userServiceClient, err := Get().GetUserServiceClient()
	if err != nil {
		return nil, err
	}
	if userServiceClient == nil {
		return nil, errors.New("user service client is nil")
	}
	resp, err := userServiceClient.GetUser(context.Background(), &v1.GetUserRequest{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	return pb2DashFunUser(resp), nil
}

// UserLogin 根据授权数据登录
// return user, isNewCreate, error
func (u *UserCenterRpc) UserLogin(authData *utils.AuthData, referrerId string, autoCreate bool) (*data.DashFunUser, bool, error) {
	userServiceClient, err := Get().GetUserServiceClient()
	if err != nil {
		return nil, false, err
	}

	ctx, cancel := NewAuthDataOutgoingContext(authData.Method, authData.Token, 5*time.Second)
	defer cancel()

	result, err := userServiceClient.Login(ctx, &v1.LoginRequest{
		ReferrerId: referrerId,
		AutoCreate: autoCreate,
	})
	if err != nil {
		return nil, false, err
	}
	return pb2DashFunUser(result.User), result.User.NewCreated, nil
}

func (u *UserCenterRpc) GetDashFunUserByAuthData(authData *utils.AuthData, onlineUserOnly bool) (*data.DashFunUser, error) {
	userServiceClient, err := Get().GetUserServiceClient()
	if err != nil {
		return nil, err
	}

	ctx, cancel := NewAuthDataOutgoingContext(authData.Method, authData.Token, 5*time.Second)
	defer cancel()

	r, err := userServiceClient.GetUserByAuth(ctx, &v1.GetUserByAuthRequest{OnlineOnly: onlineUserOnly})
	if err != nil {
		return nil, err
	}

	return pb2DashFunUser(r), err
}

func (u *UserCenterRpc) GetUserAvatar(userId string) ([]byte, error) {
	userServiceClient, err := Get().GetUserServiceClient()
	if err != nil {
		return nil, err
	}

	r, err := userServiceClient.GetUserAvatar(context.Background(), &v1.GetUserAvatarRequest{UserId: userId})
	if err != nil {
		return nil, err
	}
	return r.Avatar, nil
}

func (u *UserCenterRpc) GetUserChannelId(userId string, from data.DashFunUserFrom) (string, error) {
	userServiceClient, err := Get().GetUserServiceClient()
	if err != nil {
		return "", err
	}

	resp, err := userServiceClient.GetUserChannelId(context.Background(), &v1.GetUserChannelIdRequest{
		UserId: userId,
		From:   int32(from),
	})
	if err != nil {
		return "", err
	}
	return resp.ChannelId, nil
}

func (u *UserCenterRpc) CreateUser(from data.DashFunUserFrom, username string) (*data.DashFunUser, error) {
	userServiceClient, err := Get().GetUserServiceClient()
	if err != nil {
		return nil, err
	}

	resp, err := userServiceClient.CreateUser(context.Background(), &v1.CreateUserRequest{
		From:     int32(from),
		Username: username,
	})
	if err != nil {
		return nil, err
	}
	return pb2DashFunUser(resp.User), nil
}

func (u *UserCenterRpc) GetUsersFrom(from data.DashFunUserFrom) ([]*data.DashFunUser, error) {
	userServiceClient, err := Get().GetUserServiceClient()
	if err != nil {
		return nil, err
	}

	resp, err := userServiceClient.GetUsersFrom(context.Background(), &v1.GetUserFromRequest{
		From: int32(from),
	})
	if err != nil {
		return nil, err
	}

	users := make([]*data.DashFunUser, len(resp.Users))
	for i, user := range resp.Users {
		users[i] = pb2DashFunUser(user)
	}
	return users, nil
}

func pb2DashFunUser(user *common.DashFunUserPb) *data.DashFunUser {
	return &data.DashFunUser{
		Id:            user.UserId,
		ChannelId:     "",
		DisplayName:   user.DisplayName,
		UserName:      user.Username,
		AvatarUrl:     "",
		From:          data.DashFunUserFrom(user.From),
		CreateData:    user.CreateDate,
		LoginTime:     user.LoginTime,
		LogoffTime:    user.LogoffTime,
		WalletAddress: user.WalletAddress,
		ReferrerId:    user.ReferrerId,
	}
}
