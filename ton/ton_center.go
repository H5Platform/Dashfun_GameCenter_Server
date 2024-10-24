package ton

import (
	"context"
	"dashfun_gamecenter/config"
	"encoding/json"
	"github.com/tonkeeper/tonapi-go"
	"github.com/tonkeeper/tongo"
	"go.uber.org/zap"
	"sync"
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

	collectionAddr := "EQDcal2d9sU3xBxZ6TvBUPSkBEW_wtn7VYwuuWCPiX-z6Pyz"

	r, err := client.ExecGetMethodForBlockchainAccount(context.Background(), tonapi.ExecGetMethodForBlockchainAccountParams{
		AccountID:  collectionAddr,
		MethodName: "get_collection_data",
		Args:       nil,
	})
	if err != nil {
		zap.S().Error(err)
		return
	}

	data := &GetCollectionData{}

	err = json.Unmarshal(r.Decoded, data)
	if err != nil {
		zap.S().Error(err)
		return
	}

	addr := tongo.MustParseAddress(data.OwnerAddress)

	zap.S().Infow("result", "r", r, "r1", r.Decoded, "addr", addr.ID.ToHuman(false, true))
}
