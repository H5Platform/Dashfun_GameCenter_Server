package apperrors

import "github.com/pkg/errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserDoesNotExist = errors.New("user does not exist")

	ErrOnlineUserNotExist = newError("user does not online")

	ErrUserGameSaveDataNotExisted = newError("User Game Save Data Not Existed")

	ErrRechargeOrderCreateFailed = newError("order create failed")
	ErrRechargeOrderNotFound     = newError("order not found")
	ErrRechargeOrderStatus       = newError("order status error")
	ErrRechargeOrderCantCancel   = newError("you can't cancel this order")

	ErrUserPhotoNotExist = newError("user photo not exist")

	ErrPaymentNotEnoughBalance = newError("not enough balance")
)

func newError(msg string) error {
	return errors.New(msg)
}
