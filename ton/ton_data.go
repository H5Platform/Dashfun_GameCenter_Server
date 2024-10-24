package ton

import (
	"dashfun_gamecenter/config"
	"github.com/tonkeeper/tongo"
)

type GetCollectionData struct {
	NextItemIndex     string `json:"next_item_index"`
	CollectionContent string `json:"collection_content"`
	OwnerAddress      string `json:"owner_address"`
}

func (d *GetCollectionData) GetOwnerAddress() (string, error) {
	return toHumanAddress(d.OwnerAddress)
}

func toHumanAddress(address string) (string, error) {
	addr, err := tongo.ParseAddress(address)
	if err != nil {
		return "", err
	}
	return addr.ID.ToHuman(false, config.GetConfig().TonCfg.IsTest), nil
}
