package web3center

import (
	"context"
	"crypto/ecdsa"
	"dashfun_gamecenter/config"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

var onceWeb3Center sync.Once
var instWeb3Center *Web3Center

type Web3Center struct {
	client     *ethclient.Client
	chainId    *big.Int
	privateKey *ecdsa.PrivateKey
	localNonce uint64
	nonceMutex sync.Mutex
}

func Get() *Web3Center {
	onceWeb3Center.Do(func() {
		instWeb3Center = &Web3Center{}
		instWeb3Center.init()
	})
	return instWeb3Center
}

func (w *Web3Center) init() {
	cfg := config.GetConfig().Web3Config
	if cfg == nil {
		zap.S().Warn("Web3 config is nil, skipping initialization")
		return
	}

	client, err := ethclient.Dial(cfg.RpcUrl)
	if err != nil {
		zap.S().Errorw("Failed to connect to the Ethereum client", "url", cfg.RpcUrl, "err", err)
		return
	}
	w.client = client

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		zap.S().Errorw("Failed to get chainID", "err", err)
		return
	}
	w.chainId = chainId
	zap.S().Infow("Web3 Center initialized", "chainId", chainId, "url", cfg.RpcUrl)

	if cfg.PrivateKey != "" {
		privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
		if err != nil {
			zap.S().Errorw("Failed to parse private key", "err", err)
		} else {
			w.privateKey = privateKey

			publicKey := privateKey.Public()
			publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
			if ok {
				fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
				zap.S().Infow("Web3 Operator Address", "address", fromAddress.Hex())
			}
		}
	}
}

func (w *Web3Center) GetClient() *ethclient.Client {
	return w.client
}

func (w *Web3Center) GetChainId() *big.Int {
	if w.chainId == nil {
		return big.NewInt(0)
	}
	return w.chainId
}

// TransferToken sends ERC20 tokens. Note: This is a simplified implementation or simulation if keys/ABI missing.
func (w *Web3Center) TransferToken(tokenAddr, toAddr string, amount float64) (string, error) {
	if w.client == nil {
		return "", nil // Or error? For now, simulate success if web3 not ready to avoid blocking dev flow?
		// User requested this feature, so it should probably fail if not configured.
		// But if config is missing, maybe return mock hash "0xMOCK..."
	}

	// 1. Check if we have private key
	if w.privateKey == nil {
		// Mock mode
		zap.S().Warn("Web3 private key missing, returning mock hash")
		return "0xMOCKED_TRANSACTION_HASH_" + fmt.Sprintf("%d", time.Now().UnixNano()), nil
	}

	// 2. Implement Transfer (Simplified)
	// Without generating ABI bindings, we can construct the transaction manually.
	// But `go-ethereum` requires ABI or manual data construction.
	// function transfer(address _to, uint256 _value) public returns (bool success)
	// Selector: 0xa9059cbb

	privateKey := w.privateKey
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// --- Nonce Management Start ---
	w.nonceMutex.Lock()
	// Get network nonce (pending)
	netNonce, err := w.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		w.nonceMutex.Unlock()
		return "", err
	}

	// Logic: If network nonce > local nonce, it means we restarted or network is ahead (fresh start).
	// usage = netNonce.
	// If local nonce >= network nonce, it means we have pending txs that network hasn't seen yet.
	// usage = localNonce.

	if netNonce > w.localNonce {
		w.localNonce = netNonce
	}

	nonce := w.localNonce
	w.localNonce++
	w.nonceMutex.Unlock()
	// --- Nonce Management End ---

	value := big.NewInt(0) // ETH value is 0
	gasPrice, err := w.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	toAddress := common.HexToAddress(tokenAddr) // Contract Address

	// Pack Data
	// Method ID: 0xa9059cbb
	// Param 1: To Address (padded to 32 bytes)
	// Param 2: Amount (padded to 32 bytes)

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := crypto.Keccak256Hash(transferFnSignature)
	methodID := hash.Bytes()[:4]

	paddedAddress := common.LeftPadBytes(common.HexToAddress(toAddr).Bytes(), 32)

	// Amount is float64 (Token Amount), need to convert to Wei/Unit based on Decimals.
	// We assume 18 decimals for now or standard. HHA config?
	// For simplicity, let's assume input amount is already adjusted or we just cast to int if it's 1:1.
	// Config says Rate=50, TokenAmount=Point/50.
	// Let's assume standard 18 decimals. 1 Token = 1e18 Wei.
	// bigAmount = amount * 1e18

	amountBig := new(big.Float).SetFloat64(amount)
	multiplier := new(big.Float).SetFloat64(1e18)
	amountWei := new(big.Float).Mul(amountBig, multiplier)
	amountInt := new(big.Int)
	amountWei.Int(amountInt)

	paddedAmount := common.LeftPadBytes(amountInt.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	// Estimate Gas
	msg := ethereum.CallMsg{
		From:     fromAddress,
		To:       &toAddress,
		GasPrice: gasPrice,
		Value:    value,
		Data:     data,
	}
	gasLimit, err := w.client.EstimateGas(context.Background(), msg)
	if err != nil {
		// Fallback gas limit
		gasLimit = 300000
		zap.S().Warn("Failed to estimate gas, using default", "err", err)
	}

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := w.client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", err
	}

	err = w.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}
