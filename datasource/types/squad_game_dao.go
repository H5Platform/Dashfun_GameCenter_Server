package types

import (
	"dashfun_gamecenter/datasource/data"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SquadGameDao interface {
	SaveBet(bet *data.SquadGameBet) error
	GetUnclaimedWinningBets(roundId int64, faction int) ([]*data.SquadGameBet, error)
	GetUserUnclaimedBets(userAddress string) ([]*data.SquadGameBet, error)
	UpdateBetClaimed(betId primitive.ObjectID) error
	SaveRound(round *data.SquadGameRound) error
	GetRound(roundId int64) (*data.SquadGameRound, error)
	GetLastBlockNumber() (int64, error)
	SetLastBlockNumber(blockNumber int64) error
	GetUserLatestBet(userAddress string) (*data.SquadGameBet, error)
}
