package AirdropCenter

import (
	"context"
	"crypto/ecdsa"
	"dashfun_gamecenter/config"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"math/big"
	"os"
	"strings"
)

var vestingContractAbi string

type VestingContract struct {
	Address string `json:"address"` // 锁仓合约地址
	Abi     string `json:"abi"`
}

func NewVestingContract(address string) (*VestingContract, error) {
	if vestingContractAbi == "" {
		// Read the ABI from the file
		abiBytes, err := os.ReadFile("./conf/vesting_contract_abi.json")
		if err != nil {
			return nil, errors.Wrap(err, "failed to read vesting contract ABI file")
		}
		vestingContractAbi = string(abiBytes)
	}

	return &VestingContract{
		Address: address,
		Abi:     vestingContractAbi,
	}, nil
}

func (vc *VestingContract) GetAddress() common.Address {
	return common.HexToAddress(vc.Address)
}

// CreateVesting creates a vesting schedule for the specified recipient address with the given amount.
// If ignoreInitUnlock is true, it will not unlock the initial amount (e.g., for KuCoin users).
func (vc *VestingContract) CreateVesting(recipientAddress string, amount string, ignoreInitUnlock bool) (string, error) {
	return vc.makeCreateVestingTransaction(recipientAddress, amount, ignoreInitUnlock)
}

func (vc *VestingContract) makeCreateVestingTransaction(recipientAddress string, amount string, ignoreInitUnlock bool) (string, error) {
	client, err := ethclient.Dial(config.GetConfig().Web3Config.RpcUrl)
	if err != nil {
		return "", err
	}

	privateKey, err := crypto.HexToECDSA(config.GetConfig().Web3Config.PrivateKey)
	if err != nil {
		return "", err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 获取 nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", errors.Wrap(err, "failed to get nonce")
	}

	// 设置 gas 和链 ID
	gasLimit := uint64(300_000) // 可根据合约方法调整
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", errors.Wrap(err, "failed to suggest gas price")
	}
	chainID := big.NewInt(int64(config.GetConfig().Web3Config.ChainId))

	contractAddress := vc.GetAddress()
	parseAbi, err := abi.JSON(strings.NewReader(vc.Abi))
	if err != nil {
		return "", errors.Wrap(err, "failed to parse vesting contract ABI")
	}

	// Convert amount (string, in Ether) to *big.Int in Wei
	amountFloat, ok := new(big.Float).SetString(amount)
	if !ok {
		return "", errors.New("invalid amount string")
	}
	weiFloat := new(big.Float).Mul(amountFloat, big.NewFloat(1e18))
	wei := new(big.Int)
	weiFloat.Int(wei)

	input, err := parseAbi.Pack("createVesting", common.HexToAddress(recipientAddress), wei, ignoreInitUnlock)

	// 构造交易
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &contractAddress,
		Value:    big.NewInt(0),
		Data:     input,
	})

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)

	if err != nil {
		return "", errors.Wrap(err, "failed to sign transaction")
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", errors.Wrap(err, "failed to send transaction")
	}

	txHash := signedTx.Hash().Hex()
	return txHash, nil
}
