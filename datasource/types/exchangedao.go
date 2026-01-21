package types

import (
	"context"
	"dashfun_gamecenter/datasource/data"
)

type ExchangeDao interface {
	Save(ctx context.Context, log *data.ExchangeLog) error
	GetDailyGlobalUsage(ctx context.Context, date string) (float64, error)
	GetDailyUserUsage(ctx context.Context, userId, date string) (float64, error)
	UpdateStatus(ctx context.Context, id string, status int, txHash string) error
	GetUnissuedLogs(ctx context.Context) ([]*data.ExchangeLog, error)
	GetExchangeHistory(ctx context.Context, userId string, limit int) ([]*data.ExchangeLog, error)
}
