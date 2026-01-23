package nolandata

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var Wallets = []string{}
var PrivateKeys = []string{}
var CurrentWalletIndex = 0

type WalletState struct {
	CurrentIndex int `yaml:"current_index" json:"current_index"`
}

type WalletItem struct {
	Address    string `yaml:"address"`
	PrivateKey string `yaml:"private_key"`
}

func init() {
	LoadWallets()
	LoadWalletState()
}

func LoadWallets() {
	filePath := "conf/wallets.yml"
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to read wallets config: %v", err)
	}

	var items []WalletItem
	err = yaml.Unmarshal(data, &items)
	if err != nil {
		log.Fatalf("failed to unmarshal wallets config: %v", err)
	}

	Wallets = make([]string, 0, len(items))
	PrivateKeys = make([]string, 0, len(items))

	for _, item := range items {
		Wallets = append(Wallets, item.Address)
		PrivateKeys = append(PrivateKeys, item.PrivateKey)
	}

	log.Printf("Successfully loaded %d wallets from %s", len(Wallets), filePath)
}

func LoadWalletState() {
	filePath := "conf/bot_wallet_state.yml"
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			CurrentWalletIndex = 0
			return
		}
		log.Printf("failed to read wallet state: %v", err)
		return
	}

	var state WalletState
	err = yaml.Unmarshal(data, &state)
	if err != nil {
		log.Printf("failed to unmarshal wallet state: %v", err)
		return
	}
	CurrentWalletIndex = state.CurrentIndex
}

func SaveWalletState() {
	filePath := "conf/bot_wallet_state.yml"
	state := WalletState{
		CurrentIndex: CurrentWalletIndex,
	}
	data, err := yaml.Marshal(&state)
	if err != nil {
		log.Printf("failed to marshal wallet state: %v", err)
		return
	}
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		log.Printf("failed to write wallet state: %v", err)
	}
}

func GetNextWallet() (string, int) {
	if len(Wallets) == 0 {
		return "", -1
	}

	// Double check to be safe
	if CurrentWalletIndex >= len(Wallets) {
		CurrentWalletIndex = 0
	}

	addr := Wallets[CurrentWalletIndex]
	idx := CurrentWalletIndex

	CurrentWalletIndex++
	if CurrentWalletIndex >= len(Wallets) {
		CurrentWalletIndex = 0
	}
	SaveWalletState()

	return addr, idx
}

func GetWalletByIndex(index int) (string, string) {
	if index < 0 || index >= len(Wallets) {
		return "", ""
	}
	return Wallets[index], PrivateKeys[index]
}
