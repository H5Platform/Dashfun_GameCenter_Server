package ton

import (
	"dashfun_gamecenter/config"
	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
	"github.com/tonkeeper/tongo/ton"
)

type GetCollectionData struct {
	NextItemIndex     string `json:"next_item_index"`
	CollectionContent string `json:"collection_content"`
	OwnerAddress      string `json:"owner_address"`
}

func (d *GetCollectionData) GetOwnerAddress() (string, error) {
	return toHumanAddress(d.OwnerAddress)
}

func (d *GetCollectionData) MustGetOwnerAddress() string {
	r, err := toHumanAddress(d.OwnerAddress)
	if err != nil {
		return ""
	}
	return r
}

type MsgMintNft struct {
	Op        uint32
	QueryId   uint64
	ToAddress tlb.MsgAddress
	Quality   uint8
}

func (d *MsgMintNft) ToCell() (*boc.Cell, error) {
	c := boc.NewCell()
	err := tlb.Marshal(c, d)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewMingNftMsg(queryId uint64, toAddress ton.AccountID, quality uint8) *MsgMintNft {
	return &MsgMintNft{
		Op:        0x3e6aa798,
		QueryId:   queryId,
		ToAddress: toAddress.ToMsgAddress(),
		Quality:   quality,
	}
}

func toHumanAddress(address string) (string, error) {
	addr, err := tongo.ParseAddress(address)
	if err != nil {
		return "", err
	}
	return addr.ID.ToHuman(false, config.GetConfig().TonCfg.IsTest), nil
}
