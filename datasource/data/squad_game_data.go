package data

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SquadGameBet struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoundId     int64              `bson:"round_id" json:"round_id"`
	UserAddress string             `bson:"user_address" json:"user_address"`
	Faction     int                `bson:"faction" json:"faction"`
	Amount      string             `bson:"amount" json:"amount"` // Stored as string to preserve precision
	TxHash      string             `bson:"tx_hash" json:"tx_hash"`
	IsClaimed   bool               `bson:"is_claimed" json:"is_claimed"`
	CreatedAt   int64              `bson:"created_at" json:"created_at"`
}

type SquadGameRound struct {
	Id                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoundId            int64              `bson:"round_id" json:"round_id"`
	WinningFaction     int                `bson:"winning_faction" json:"winning_faction"`
	TotalPool          string             `bson:"total_pool" json:"total_pool"`
	WinnerPool         string             `bson:"winner_pool" json:"winner_pool"`
	FactionPools       []string           `bson:"faction_pools" json:"faction_pools"` // Amount for each faction
	SettledBlockNumber int64              `bson:"settled_block_number" json:"settled_block_number"`
	SettledBlockHash   string             `bson:"settled_block_hash" json:"settled_block_hash"`
	IsSettled          bool               `bson:"is_settled" json:"is_settled"`
	UpdatedAt          int64              `bson:"updated_at" json:"updated_at"`
}

type SquadGameGameState struct {
	Id              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Key             string             `bson:"key" json:"key"` // e.g. "scanner_state"
	LastBlockNumber int64              `bson:"last_block_number" json:"last_block_number"`
	UpdatedAt       int64              `bson:"updated_at" json:"updated_at"`
}

type SquadGameUserInfo struct {
	Bet          *SquadGameBet   `json:"bet"`
	Round        *SquadGameRound `json:"round"`
	IsGameActive bool            `json:"is_game_active"`
}
