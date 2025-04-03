package ton

import (
	"context"
	"dashfun_gamecenter/config"
	"encoding/json"
	"github.com/tonkeeper/tonapi-go"
	"github.com/tonkeeper/tongo/liteapi"
	"github.com/tonkeeper/tongo/ton"
	"github.com/tonkeeper/tongo/wallet"
	"go.uber.org/zap"
	"sync"
	"time"
)

var onceTonCenter sync.Once
var instTonCenter *TonCenter

type TonCenter struct {
	client *tonapi.Client
}

func Get() *TonCenter {
	onceTonCenter.Do(func() {
		instTonCenter = &TonCenter{}
		instTonCenter.init()
	})
	return instTonCenter
}

func (t *TonCenter) init() {
	url := tonapi.TonApiURL
	if config.GetConfig().TonCfg.IsTest {
		url = tonapi.TestnetTonApiURL
	}
	client, err := tonapi.NewClient(url, tonapi.WithToken(config.GetConfig().TonCfg.ApiKey))
	if err != nil {
		panic(err)
	}
	t.client = client
	t.SendMsgTest()
}

func (t *TonCenter) getCollectionData() (*GetCollectionData, error) {
	collectionAddr := "EQDcal2d9sU3xBxZ6TvBUPSkBEW_wtn7VYwuuWCPiX-z6Pyz"

	r, err := t.client.ExecGetMethodForBlockchainAccount(context.Background(), tonapi.ExecGetMethodForBlockchainAccountParams{
		AccountID:  collectionAddr,
		MethodName: "get_collection_data",
		Args:       nil,
	})
	if err != nil {
		zap.S().Error(err)
		return nil, err
	}

	data := &GetCollectionData{}

	err = json.Unmarshal(r.Decoded, data)
	if err != nil {
		zap.S().Error(err)
		return nil, err
	}
	zap.S().Infow("result", "r", r, "r1", r.Decoded, "addr", data.MustGetOwnerAddress())
	return data, nil
}

func (t *TonCenter) SendMsgTest() {
	collectionAddr := "EQDcal2d9sU3xBxZ6TvBUPSkBEW_wtn7VYwuuWCPiX-z6Pyz"
	collectionId := ton.MustParseAccountID(collectionAddr)
	seed := config.GetConfig().TonCfg.WalletMnemonic
	pk, err := wallet.SeedToPrivateKey(seed)
	if err != nil {
		zap.S().Error(err)
		return
	}
	cli, err := liteapi.NewClient(liteapi.Testnet())
	if err != nil {
		zap.S().Error(err)
		return
	}

	networkId, err := cli.GetNetworkGlobalID(context.Background())
	if err != nil {
		zap.S().Error(err)
		return
	}

	v := t.getWalletVersion(config.GetConfig().TonCfg.WalletVersion)

	w, err := wallet.New(pk, v, cli, wallet.WithNetworkGlobalID(networkId))
	if err != nil {
		zap.S().Error(err)
		return
	}
	zap.S().Infow(w.GetAddress().ToHuman(false, true))

	add := w.GetAddress()

	body, err := NewMingNftMsg(uint64(time.Now().UnixMilli()), add, 1).ToCell()

	if err != nil {
		zap.S().Error(err)
		return
	}

	v2, err := w.SendV2(context.Background(), 0, wallet.Message{
		Amount:  ton.OneTON / 10,
		Address: collectionId,
		Body:    body,
	})
	if err != nil {
		zap.S().Error(err)
		return
	}

	zap.S().Infow("v2", "v2", v2.Hex())
}

func (t *TonCenter) getWalletVersion(verString string) wallet.Version {
	switch verString {
	case "V4", "V4R2":
		return wallet.V4R2
	case "V3", "V3R2":
		return wallet.V3R2
	}
	return wallet.V4R2
}
