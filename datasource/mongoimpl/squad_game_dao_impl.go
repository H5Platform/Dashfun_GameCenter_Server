package mongoimpl

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/datasource/types"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	squadGameDao     *SquadGameDaoMongo
	squadGameDaoOnce sync.Once
)

type SquadGameDaoMongo struct {
	betCollection   *mongo.Collection
	roundCollection *mongo.Collection
	stateCollection *mongo.Collection
}

func GetSquadGameDaoMongo() types.SquadGameDao {
	squadGameDaoOnce.Do(func() {
		db := GetMongoDatabase()
		squadGameDao = &SquadGameDaoMongo{
			betCollection:   db.Collection("squad_game_bets"),
			roundCollection: db.Collection("squad_game_rounds"),
			stateCollection: db.Collection("squad_game_state"),
		}
		// Initialize indexes
		CreateIndexes(squadGameDao.betCollection, []IndexInfo{
			{FieldName: "round_id", Unique: false, Sort: 1, IndexName: "idx_round_id"},
			{FieldName: "user_address", Unique: false, Sort: 1, IndexName: "idx_user_address"},
			{FieldName: "faction", Unique: false, Sort: 1, IndexName: "idx_faction"},
			{FieldName: "created_at", Unique: false, Sort: 1, IndexName: "idx_created_at"},
		})
		CreateIndexes(squadGameDao.roundCollection, []IndexInfo{
			{FieldName: "round_id", Unique: true, Sort: 1, IndexName: "idx_round_id_unique"},
		})
	})
	return squadGameDao
}

func (d *SquadGameDaoMongo) SaveBet(bet *data.SquadGameBet) error {
	if bet.Id.IsZero() {
		bet.Id = primitive.NewObjectID()
		bet.CreatedAt = time.Now().Unix()
	}
	_, err := d.betCollection.InsertOne(context.TODO(), bet)
	return err
}

func (d *SquadGameDaoMongo) GetUnclaimedWinningBets(roundId int64, faction int) ([]*data.SquadGameBet, error) {
	filter := bson.D{
		{Key: "round_id", Value: roundId},
		{Key: "faction", Value: faction},
		{Key: "is_claimed", Value: false},
	}
	cursor, err := d.betCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var bets []*data.SquadGameBet
	if err = cursor.All(context.TODO(), &bets); err != nil {
		return nil, err
	}
	return bets, nil
}

func (d *SquadGameDaoMongo) GetUserUnclaimedBets(userAddress string) ([]*data.SquadGameBet, error) {
	filter := bson.D{
		{Key: "user_address", Value: userAddress},
		{Key: "is_claimed", Value: false},
	}
	cursor, err := d.betCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var bets []*data.SquadGameBet
	if err = cursor.All(context.TODO(), &bets); err != nil {
		return nil, err
	}
	return bets, nil
}

func (d *SquadGameDaoMongo) UpdateBetClaimed(betId primitive.ObjectID) error {
	filter := bson.D{{Key: "_id", Value: betId}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "is_claimed", Value: true}}}}
	_, err := d.betCollection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (d *SquadGameDaoMongo) SaveRound(round *data.SquadGameRound) error {
	// Upsert based on RoundId
	filter := bson.D{{Key: "round_id", Value: round.RoundId}}
	if round.Id.IsZero() {
		// New record might be inserted, let's look if it exists or define defaults
		round.UpdatedAt = time.Now().Unix()
	}

	opt := options.Update().SetUpsert(true)
	update := bson.D{{Key: "$set", Value: round}}

	// If it's an insert, we want to ensure _id is generated if not provided,
	// but mongo driver handles _id generation on insert.
	// However, for pure InsertOne we do it manually above. For UpdateOne with Upsert,
	// we rely on mongo, but if we want to retrieve it later we might need to query by RoundId.

	_, err := d.roundCollection.UpdateOne(context.TODO(), filter, update, opt)
	return err
}

func (d *SquadGameDaoMongo) GetRound(roundId int64) (*data.SquadGameRound, error) {
	filter := bson.D{{Key: "round_id", Value: roundId}}
	var round data.SquadGameRound
	err := d.roundCollection.FindOne(context.TODO(), filter).Decode(&round)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Return nil if not found
		}
		return nil, err
	}
	return &round, nil
}

const ScannerStateKey = "scanner_state"

func (d *SquadGameDaoMongo) GetLastBlockNumber() (int64, error) {
	filter := bson.D{{Key: "key", Value: ScannerStateKey}}
	var state data.SquadGameGameState
	err := d.stateCollection.FindOne(context.TODO(), filter).Decode(&state)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil
		}
		return 0, err
	}
	return state.LastBlockNumber, nil
}

func (d *SquadGameDaoMongo) SetLastBlockNumber(blockNumber int64) error {
	filter := bson.D{{Key: "key", Value: ScannerStateKey}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "last_block_number", Value: blockNumber},
		{Key: "updated_at", Value: time.Now().Unix()},
	}}}
	opt := options.Update().SetUpsert(true)
	_, err := d.stateCollection.UpdateOne(context.TODO(), filter, update, opt)
	return err
}

func (d *SquadGameDaoMongo) GetUserLatestBet(userAddress string) (*data.SquadGameBet, error) {
	filter := bson.D{{Key: "user_address", Value: userAddress}}
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var bet data.SquadGameBet
	err := d.betCollection.FindOne(context.TODO(), filter, opts).Decode(&bet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &bet, nil
}
