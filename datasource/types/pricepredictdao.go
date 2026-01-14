package types

import (
	"context"
	"dashfun_gamecenter/datasource/data"
)

type PricePredictDao interface {
	SaveOrUpdate(data *data.PricePredictData) error
	FindById(id string) (*data.PricePredictData, error)
	FindByUserAndDate(userId, date string) (*data.PricePredictData, error)
	FindLatestByUser(userId string) (*data.PricePredictData, error)
	FindPendingPredictionsByDate(date string) ([]*data.PricePredictData, error)
	FindBots(ctx context.Context, limit int64) ([]*data.PricePredictData, error)
}
