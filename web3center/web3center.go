package web3center

import (
	"context"
	"crypto/ecdsa"
	"dashfun_gamecenter/config"
	"math/big"
	"sync"

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
