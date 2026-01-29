package squadgame

import (
	"context"
	"crypto/ecdsa"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/datasource/types"
	game "dashfun_gamecenter/nolandev/squadgame/binding"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"dashfun_gamecenter/nolandev/nolandata"
	"math/rand"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ... (existing code)

var (
	squadGameService *SquadGameService
	once             sync.Once
)

type SquadGameService struct {
	client          *ethclient.Client
	game            *game.HourlySquadGame
	auth            *bind.TransactOpts
	contractAddress common.Address
	dao             types.SquadGameDao

	nonceMutex   sync.Mutex
	currentNonce uint64
}

func GetSquadGameService() *SquadGameService {
	once.Do(func() {
		cfg := config.GetConfig()
		if cfg.HourlySquadGameCfg == nil {
			zap.S().Info("HourlySquadGameCfg is missing")
			return
		}

		squadGameConfig := cfg.HourlySquadGameCfg
		if squadGameConfig.ContractAddress == "" {
			zap.S().Info("HourlySquadGame ContractAddress is empty")
			return
		}

		client, err := ethclient.Dial(cfg.Web3Config.RpcUrl)
		if err != nil {
			zap.S().Fatalf("Failed to connect to the Ethereum client: %v", err)
		}

		address := common.HexToAddress(squadGameConfig.ContractAddress)
		instance, err := game.NewHourlySquadGame(address, client)
		if err != nil {
			zap.S().Fatalf("Failed to instantiate contract: %v", err)
		}

		// Setup Auth
		privateKeyStr := squadGameConfig.AdminPrivateKey
		if privateKeyStr == "" {
			// Fallback to main web3 key if specific admin key not provided
			privateKeyStr = cfg.Web3Config.PrivateKey
		}

		// Remove 0x prefix if present
		if len(privateKeyStr) > 2 && privateKeyStr[:2] == "0x" {
			privateKeyStr = privateKeyStr[2:]
		}

		privateKey, err := crypto.HexToECDSA(privateKeyStr)
		if err != nil {
			zap.S().Fatalf("Failed to load private key: %v", err)
		}

		chainId := big.NewInt(int64(cfg.Web3Config.ChainId))
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
		if err != nil {
			zap.S().Fatalf("Failed to create transactor: %v", err)
		}

		zap.S().Infof("SquadGame Admin Address: %s", auth.From.Hex())

		// Verify Admin is Owner to avoid 0x118cdaa7 (OwnableUnauthorizedAccount)
		owner, err := instance.Owner(nil)
		if err != nil {
			zap.S().Errorf("Failed to fetch contract owner: %v", err)
		} else {
			if owner != auth.From {
				zap.S().Warnf("CRITICAL CONFIG WARNING: Admin Address (%s) is NOT the Contract Owner (%s). DistributeRewards will fail.", auth.From.Hex(), owner.Hex())
			} else {
				zap.S().Infof("Verified: Admin is Contract Owner.")
			}
		}

		// Init Nonce
		nonce, err := client.PendingNonceAt(context.Background(), auth.From)
		if err != nil {
			zap.S().Fatalf("Failed to retrieve account nonce: %v", err)
		}

		squadGameService = &SquadGameService{
			client:          client,
			game:            instance,
			auth:            auth,
			contractAddress: address,
			dao:             dao.GetSquadGameDao(),
			currentNonce:    nonce,
		}

		// Start background tasks
		go squadGameService.startEventListeners()
		go squadGameService.startScheduler()
	})
	return squadGameService
}

func (s *SquadGameService) startEventListeners() {
	// Start polling loop
	go func() {
		// Initialize starting block
		startBlock := uint64(0)

		// Try load from DB
		lastBlock, err := s.dao.GetLastBlockNumber()
		if err != nil {
			zap.S().Infof("Failed to get last block from DB: %v", err)
		}
		if lastBlock > 0 {
			startBlock = uint64(lastBlock) + 1
			zap.S().Infof("Resuming event scanning from block %d", startBlock)
		} else {
			// Fallback to latest
			header, err := s.client.HeaderByNumber(context.Background(), nil)
			if err != nil {
				zap.S().Infof("Failed to get latest header: %v", err)
				return
			}
			startBlock = header.Number.Uint64()
			zap.S().Infof("Starting event scanning from latest block %d", startBlock)
		}

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Get latest block
			header, err := s.client.HeaderByNumber(context.Background(), nil)
			if err != nil {
				zap.S().Infof("Failed to get latest header: %v", err)
				continue
			}
			endBlock := header.Number.Uint64()

			if endBlock < startBlock {
				continue
			}

			// Cap range to avoid huge queries if service was down/stuck (e.g. max 2000 blocks)
			if endBlock-startBlock > 2000 {
				endBlock = startBlock + 2000
			}

			// Filter BetPlaced
			s.pollBetPlaced(startBlock, endBlock)

			// Filter RoundSettled
			s.pollRoundSettled(startBlock, endBlock)

			// Provide persistence
			if err := s.dao.SetLastBlockNumber(int64(endBlock)); err != nil {
				zap.S().Infof("Failed to save last block number: %v", err)
			}

			// Move startBlock pointer
			startBlock = endBlock + 1
		}
	}()
}

func (s *SquadGameService) pollBetPlaced(start, end uint64) {
	filterOpts := &bind.FilterOpts{
		Start:   start,
		End:     &end,
		Context: context.Background(),
	}

	iter, err := s.game.FilterBetPlaced(filterOpts, nil, nil)
	if err != nil {
		zap.S().Infof("Failed to filter BetPlaced: %v", err)
		return
	}
	defer iter.Close()

	for iter.Next() {
		s.handleBetPlaced(iter.Event)
	}
	if iter.Error() != nil {
		zap.S().Infof("BetPlaced iterator error: %v", iter.Error())
	}
}

func (s *SquadGameService) pollRoundSettled(start, end uint64) {
	filterOpts := &bind.FilterOpts{
		Start:   start,
		End:     &end,
		Context: context.Background(),
	}

	iter, err := s.game.FilterRoundSettled(filterOpts, nil)
	if err != nil {
		zap.S().Infof("Failed to filter RoundSettled: %v", err)
		return
	}
	defer iter.Close()

	for iter.Next() {
		s.handleRoundSettled(iter.Event)
	}
	if iter.Error() != nil {
		zap.S().Infof("RoundSettled iterator error: %v", iter.Error())
	}
}

func (s *SquadGameService) handleBetPlaced(event *game.HourlySquadGameBetPlaced) {
	zap.S().Infof("BetPlaced: Round %d, User %s, Faction %d, Amount %s",
		event.RoundId, event.User.Hex(), event.Faction, event.Amount.String())

	bet := &data.SquadGameBet{
		RoundId:     event.RoundId.Int64(),
		UserAddress: strings.ToLower(event.User.Hex()),
		Faction:     int(event.Faction),
		Amount:      event.Amount.String(),
		TxHash:      event.Raw.TxHash.Hex(),
		IsClaimed:   false,
		CreatedAt:   time.Now().Unix(),
	}

	if err := s.dao.SaveBet(bet); err != nil {
		zap.S().Infof("Error saving bet: %v", err)
	}
}

func (s *SquadGameService) handleRoundSettled(event *game.HourlySquadGameRoundSettled) {
	zap.S().Infof("RoundSettled: Round %d, Winner %d", event.RoundId, event.WinningFaction)

	factionPools := make([]string, len(event.FactionPools))
	for i, pool := range event.FactionPools {
		factionPools[i] = pool.String()
	}

	round := &data.SquadGameRound{
		RoundId:            event.RoundId.Int64(),
		WinningFaction:     int(event.WinningFaction),
		TotalPool:          event.TotalPool.String(),
		WinnerPool:         event.WinnerPool.String(),
		SettledBlockNumber: event.SettledBlockNumber.Int64(),
		SettledBlockHash:   common.BytesToHash(event.SettledBlockHash[:]).Hex(),
		IsSettled:          true,
		FactionPools:       factionPools,
		UpdatedAt:          time.Now().Unix(),
	}

	if err := s.dao.SaveRound(round); err != nil {
		zap.S().Infof("Error saving round: %v", err)
	}

	// Trigger Reward Distribution
	// go s.DistributeRewards(round.RoundId, round.WinningFaction) // Removed: Client triggered
}

func (s *SquadGameService) startScheduler() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.TrySettleRound()
	}
}

func (s *SquadGameService) TrySettleRound() {
	now := time.Now().Unix()
	// Round duration is 3600 seconds
	currentRoundStart := (now / 3600) * 3600
	// currentRoundId := big.NewInt(currentRoundStart) // Unused
	prevRoundId := big.NewInt(currentRoundStart - 3600)

	// Check if previous round needs settlement
	// Usually checking specific hour? Or just previous ID.

	// We can check contract state for previous round
	roundInfo, err := s.game.Rounds(nil, prevRoundId)
	if err != nil {
		zap.S().Infof("Error checking round info: %v", err)
		return
	}

	if !roundInfo.IsSettled {
		zap.S().Infof("Attempting to settle round %s", prevRoundId.String())

		// Refresh nonce/gas price if needed, but go-ethereum handles nonce.
		// Gas price is suggested by client.

		tx, err := s.Transact(func(opts *bind.TransactOpts) (*ethtypes.Transaction, error) {
			return s.game.SettleRound(opts, prevRoundId)
		})
		if err != nil {
			zap.S().Infof("Failed to settle round: %v", err)
			return
		}
		zap.S().Infof("SettleRound sent: %s", tx.Hash().Hex())
	}
}

func (s *SquadGameService) ClaimUserRewards(userAddress string) (string, error) {
	zap.S().Infof("Claiming rewards for User %s", userAddress)
	userAddr := common.HexToAddress(userAddress)

	// 1. Get user's unclaimed bets
	bets, err := s.dao.GetUserUnclaimedBets(strings.ToLower(userAddress))
	if err != nil {
		return "", err
	}

	if len(bets) == 0 {
		return "", fmt.Errorf("no unclaimed bets found")
	}

	// Only process the first unclaimed bet (enforcing one-at-a-time flow)
	bet := bets[0]

	// 2. Check round status
	round, err := s.dao.GetRound(bet.RoundId)
	if err != nil {
		return "", fmt.Errorf("error getting round info: %v", err)
	}
	if round == nil {
		return "", fmt.Errorf("round %d info not found", bet.RoundId)
	}

	if !round.IsSettled {
		return "", fmt.Errorf("round %d is not settled yet", bet.RoundId)
	}

	// 3. Check Win/Loss
	if round.WinningFaction == bet.Faction {
		// Winner: Call contract to distribute
		zap.S().Infof("User %s won round %d (Faction %d). Distributing rewards...", userAddress, bet.RoundId, bet.Faction)

		winningRoundIds := []*big.Int{big.NewInt(bet.RoundId)}

		tx, err := s.Transact(func(opts *bind.TransactOpts) (*ethtypes.Transaction, error) {
			return s.game.DistributeRewards(opts, userAddr, winningRoundIds)
		})
		if err != nil {
			zap.S().Infof("Failed to distribute to %s: %v", userAddress, err)
			return "", err
		}

		zap.S().Infof("DistributeRewards tx sent for %s: %s", userAddress, tx.Hash().Hex())

		// Update DB
		if err := s.dao.UpdateBetClaimed(bet.Id); err != nil {
			zap.S().Infof("Failed to mark bet as claimed for %s: %v", userAddress, err)
		}

		return tx.Hash().Hex(), nil
	} else {
		// Loser: Just mark as claimed so they can bet again
		zap.S().Infof("User %s lost round %d. Marking as claimed.", userAddress, bet.RoundId)
		if err := s.dao.UpdateBetClaimed(bet.Id); err != nil {
			zap.S().Infof("Failed to mark bet as claimed for %s: %v", userAddress, err)
			return "", err // If DB update fails, return error so they can retry
		}
		return "claimed_no_reward", nil
	}
}

// Helper to relay meta-transactions
type UserBetRequest struct {
	Faction     uint8
	Amount      *big.Int
	Deadline    *big.Int
	V           uint8
	R           [32]byte
	S           [32]byte
	FromAddress string
}

func (s *SquadGameService) RelayBet(req UserBetRequest) (string, error) {
	// 1. Check if user accepts new bet (must have no unclaimed bets)
	if req.FromAddress != "" {
		unclaimed, err := s.dao.GetUserUnclaimedBets(strings.ToLower(req.FromAddress))
		if err != nil {
			return "", fmt.Errorf("failed to check existing bets: %v", err)
		}
		if len(unclaimed) > 0 {
			return "", fmt.Errorf("you have an unsettled or unclaimed bet. please claim it first")
		}
	}

	tx, err := s.Transact(func(opts *bind.TransactOpts) (*ethtypes.Transaction, error) {
		return s.game.PlaceBetWithPermit(opts, common.HexToAddress(req.FromAddress), req.Faction, req.Amount, req.Deadline, req.V, req.R, req.S)
	})
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

// Transact executes a transaction function safely with nonce management
func (s *SquadGameService) Transact(txFunc func(*bind.TransactOpts) (*ethtypes.Transaction, error)) (*ethtypes.Transaction, error) {
	s.nonceMutex.Lock()
	defer s.nonceMutex.Unlock()

	// Determine gas price
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %v", err)
	}

	// Create a copy of auth options to avoid concurrent modification issues on the base object
	opts := *s.auth
	opts.Nonce = big.NewInt(int64(s.currentNonce))
	opts.GasPrice = gasPrice
	opts.Context = context.Background()

	tx, err := txFunc(&opts)
	if err != nil {
		// On error, re-sync nonce to ensure we are up to date for next attempt
		if newNonce, nErr := s.client.PendingNonceAt(context.Background(), s.auth.From); nErr == nil {
			s.currentNonce = newNonce
		}
		return nil, err
	}

	// Success
	s.currentNonce++
	return tx, nil
}

// GetUserBetInfo returns bet info with round status.
// If no bet found, returns current round info.
func (s *SquadGameService) GetUserBetInfo(userAddress string) (*data.SquadGameUserInfo, error) {
	// Check if game is globally active
	cfg := config.GetConfig().HourlySquadGameCfg
	isGameActive := false
	if cfg != nil {
		isGameActive = cfg.Open
	}

	// 1. Check unclaimed
	// We prioritize unclaimed bets ("active" or "unclaimed winnings").
	var activeBet *data.SquadGameBet

	unclaimed, err := s.dao.GetUserUnclaimedBets(strings.ToLower(userAddress))
	if err != nil {
		return nil, fmt.Errorf("failed to check unclaimed: %v", err)
	}

	if len(unclaimed) > 0 {
		activeBet = unclaimed[0]
	}

	res := &data.SquadGameUserInfo{
		IsGameActive: isGameActive,
	}

	if activeBet != nil {
		res.Bet = activeBet
		// Get round info for this bet
		// Try DB first
		round, err := s.dao.GetRound(activeBet.RoundId)
		if err != nil {
			zap.S().Infof("Error getting round %d from DB: %v", activeBet.RoundId, err)
		}

		if round != nil {
			res.Round = round
		} else {
			// If not in DB (maybe not settled yet or not synced), try contract
			zap.S().Infof("Round %d not found in DB, fetching from contract...", activeBet.RoundId)
			rInfo, err := s.game.GetRoundState(nil, big.NewInt(activeBet.RoundId))
			if err != nil {
				return nil, fmt.Errorf("failed to fetch round %d from contract: %v", activeBet.RoundId, err)
			}

			factionPools := make([]string, len(rInfo.FactionPools))
			for i, pool := range rInfo.FactionPools {
				factionPools[i] = pool.String()
			}

			res.Round = &data.SquadGameRound{
				RoundId:        rInfo.RoundId.Int64(),
				WinningFaction: int(rInfo.WinningFaction),
				TotalPool:      rInfo.TotalPool.String(),
				WinnerPool:     rInfo.WinnerPool.String(),
				IsSettled:      rInfo.IsSettled,
				FactionPools:   factionPools,
				// Other fields might be empty if fetched from contract directly without event
			}
		}
	} else {
		// No active bet. Return latest round info.
		// User wants "latest round info".

		// Use GetRoundState(0) to fetch current/latest round info
		rInfo, err := s.game.GetRoundState(nil, big.NewInt(0))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch current round state: %v", err)
		}

		factionPools := make([]string, len(rInfo.FactionPools))
		for i, pool := range rInfo.FactionPools {
			factionPools[i] = pool.String()
		}

		res.Round = &data.SquadGameRound{
			RoundId:        rInfo.RoundId.Int64(),
			WinningFaction: int(rInfo.WinningFaction),
			TotalPool:      rInfo.TotalPool.String(),
			WinnerPool:     rInfo.WinnerPool.String(),
			IsSettled:      rInfo.IsSettled,
			FactionPools:   factionPools,
			// SettledBlockNumber etc might be 0 if not settled
		}
	}

	return res, nil
}

func (s *SquadGameService) GetLatestRound() (*data.SquadGameRound, error) {
	// Use GetRoundState(0) to fetch current/latest round info
	rInfo, err := s.game.GetRoundState(nil, big.NewInt(0))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current round state: %v", err)
	}

	factionPools := make([]string, len(rInfo.FactionPools))
	for i, pool := range rInfo.FactionPools {
		factionPools[i] = pool.String()
	}

	return &data.SquadGameRound{
		RoundId:        rInfo.RoundId.Int64(),
		WinningFaction: int(rInfo.WinningFaction),
		TotalPool:      rInfo.TotalPool.String(),
		WinnerPool:     rInfo.WinnerPool.String(),
		IsSettled:      rInfo.IsSettled,
		FactionPools:   factionPools,
	}, nil
}

func (s *SquadGameService) BotAutoPlay(bot *data.NolanBotData) error {
	// 0. Check Global Switch
	cfg := config.GetConfig().HourlySquadGameCfg
	if cfg == nil || !cfg.Open {
		// Game disabled, bot should not play
		return nil
	}

	// 0.5 Check Time (Don't bet if > 58th minute)
	if time.Now().Minute() >= 58 {
		zap.S().Infof("Bot %s skipping bet: too close to end of hour (minute %d)", bot.Name, time.Now().Minute())
		return nil
	}

	// 1. Check Rewards (via DB unclaimed bets)
	unclaimed, err := s.dao.GetUserUnclaimedBets(strings.ToLower(bot.WalletAddr))
	if err == nil && len(unclaimed) > 0 {
		zap.S().Infof("Bot %s has %d unclaimed bets. Claiming...", bot.Name, len(unclaimed))
		if _, err := s.ClaimUserRewards(bot.WalletAddr); err != nil {
			zap.S().Infof("Failed to claim rewards for bot %s: %v", bot.Name, err)
		} else {
			time.Sleep(2 * time.Second)
		}
	}

	// 2. Check Balance
	balance, err := s.getTokenBalance(bot.WalletAddr)
	if err != nil {
		return fmt.Errorf("failed to get bot balance: %v", err)
	}

	// 3. Determine Bet Amount
	// Options: 10, 20, 50, 100
	amounts := []float64{10, 20, 50, 100}

	// Convert balance to float for comparison (assuming 18 decimals, 1 Token = 1e18 Wei)
	balanceEth := new(big.Float).SetInt(balance)
	balanceEth.Quo(balanceEth, big.NewFloat(1e18))
	balanceFloat, _ := balanceEth.Float64()

	viable := []float64{}
	for _, a := range amounts {
		if balanceFloat >= a {
			viable = append(viable, a)
		}
	}

	if len(viable) == 0 {
		return fmt.Errorf("bot %s insufficient balance: %f", bot.Name, balanceFloat)
	}

	amountVal := viable[rand.Intn(len(viable))]

	// 4. Determine Faction
	// Hex tail of timestamp logic
	// Map: 0,4,8,12 -> 0 => val % 4 == 0.
	timestamp := time.Now().UnixMilli()
	hexTail := timestamp % 16
	faction := uint8(hexTail % 4)

	// 5. Sign Permit and Place Bet
	// Need Private Key
	_, privateKeyHex := nolandata.GetWalletByIndex(bot.WalletIndex)
	if privateKeyHex == "" {
		return fmt.Errorf("bot private key not found index %d", bot.WalletIndex)
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return fmt.Errorf("invalid private key: %v", err)
	}

	// Amount BigInt
	amountWei := new(big.Int)
	amountWei.SetString(fmt.Sprintf("%.0f000000000000000000", amountVal), 10) // 1e18

	// Deadline
	deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())

	// Nonce
	tokenAddr, _ := s.game.PaymentToken(nil)
	nonce, err := s.getTokenNonce(bot.WalletAddr, tokenAddr)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %v", err)
	}

	// Sign
	v, r, sVals, err := s.signPermit(bot.WalletAddr, tokenAddr, s.contractAddress, amountWei, nonce, deadline, privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign permit: %v", err)
	}

	// Relay
	req := UserBetRequest{
		Faction:     faction,
		Amount:      amountWei,
		Deadline:    deadline,
		V:           v,
		R:           r,
		S:           sVals,
		FromAddress: bot.WalletAddr,
	}

	zap.S().Infof("Bot %s placing bet: Faction=%d, Amount=%f", bot.Name, faction, amountVal)
	_, err = s.RelayBet(req)
	return err
}

func (s *SquadGameService) getTokenBalance(addr string) (*big.Int, error) {
	tokenAddr, err := s.game.PaymentToken(nil)
	if err != nil {
		return nil, err
	}

	data := append(common.FromHex("70a08231"), common.LeftPadBytes(common.HexToAddress(addr).Bytes(), 32)...)
	res, err := s.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &tokenAddr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	balance := new(big.Int)
	balance.SetBytes(res)
	return balance, nil
}

func (s *SquadGameService) getTokenNonce(owner string, tokenAddr common.Address) (*big.Int, error) {
	// nonces(address) selector 7ecebe00
	data := append(common.FromHex("7ecebe00"), common.LeftPadBytes(common.HexToAddress(owner).Bytes(), 32)...)
	res, err := s.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &tokenAddr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}
	nonce := new(big.Int)
	nonce.SetBytes(res)
	return nonce, nil
}

func (s *SquadGameService) signPermit(ownerStr string, tokenAddr, spenderAddr common.Address, value, nonce, deadline *big.Int, pk *ecdsa.PrivateKey) (uint8, [32]byte, [32]byte, error) {
	chainID, err := s.client.ChainID(context.Background())
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}

	// 1. Build Domain Separator
	// keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)")
	domainTypeHash := crypto.Keccak256([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))

	// Name: "SquadGameToken" (Assumption)
	nameHash := crypto.Keccak256([]byte("Harry Howard AI"))
	versionHash := crypto.Keccak256([]byte("1"))

	domainData := make([]byte, 0)
	domainData = append(domainData, domainTypeHash...)
	domainData = append(domainData, nameHash...)
	domainData = append(domainData, versionHash...)
	domainData = append(domainData, common.LeftPadBytes(chainID.Bytes(), 32)...)
	domainData = append(domainData, common.LeftPadBytes(tokenAddr.Bytes(), 32)...) // Verifying Contract is TOKEN

	domainSeparator := crypto.Keccak256(domainData)

	// 2. Build Struct Hash
	// keccak256("Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)")
	permitTypeHash := crypto.Keccak256([]byte("Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"))

	structData := make([]byte, 0)
	structData = append(structData, permitTypeHash...)
	structData = append(structData, common.LeftPadBytes(common.HexToAddress(ownerStr).Bytes(), 32)...)
	structData = append(structData, common.LeftPadBytes(spenderAddr.Bytes(), 32)...)
	structData = append(structData, common.LeftPadBytes(value.Bytes(), 32)...)
	structData = append(structData, common.LeftPadBytes(nonce.Bytes(), 32)...)
	structData = append(structData, common.LeftPadBytes(deadline.Bytes(), 32)...)

	structHash := crypto.Keccak256(structData)

	// 3. Digest
	// keccak256("\x19\x01" + domainSeparator + structHash)
	prefix := []byte{0x19, 0x01}
	digestData := append(prefix, domainSeparator...)
	digestData = append(digestData, structHash...)
	digest := crypto.Keccak256(digestData)

	// 4. Sign
	sig, err := crypto.Sign(digest, pk)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}

	// Extract v, r, s
	// Sig is [R || S || V] 65 bytes
	if len(sig) != 65 {
		return 0, [32]byte{}, [32]byte{}, fmt.Errorf("invalid signature length: %d", len(sig))
	}

	var r, sVal [32]byte
	copy(r[:], sig[:32])
	copy(sVal[:], sig[32:64])
	v := sig[64] + 27 // Transform V from 0/1 to 27/28

	return v, r, sVal, nil
}
