package apperrors

import "github.com/pkg/errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserDoesNotExist = errors.New("user does not exist")

	ErrOnlineUserNotExist = newError("user does not online")

	ErrUserGameSaveDataNotExisted = newError("User Game Save Data Not Existed")
)

func newError(msg string) error {
	return errors.New(msg)
}
