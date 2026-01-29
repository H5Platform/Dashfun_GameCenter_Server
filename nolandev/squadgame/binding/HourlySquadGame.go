// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package game

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// HourlySquadGameMetaData contains all meta data concerning the HourlySquadGame contract.
var HourlySquadGameMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_paymentToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_trustedForwarder\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"FEE_BASIS_POINTS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"LOCK_WINDOW\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ROUND_DURATION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"bets\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"faction\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"isClaimed\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"claimRewards\",\"inputs\":[{\"name\":\"roundIds\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"distributeRewards\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"roundIds\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getCurrentRoundId\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoundState\",\"inputs\":[{\"name\":\"_roundId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"roundId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"totalPool\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"factionPools\",\"type\":\"uint256[4]\",\"internalType\":\"uint256[4]\"},{\"name\":\"isOpen\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"isSettled\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"winningFaction\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"winnerPool\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"settledBlockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"settledBlockHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isBettingOpen\",\"inputs\":[{\"name\":\"timestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"isTrustedForwarder\",\"inputs\":[{\"name\":\"forwarder\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxBetAmount\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"paymentToken\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIERC20\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"placeBet\",\"inputs\":[{\"name\":\"faction\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"placeBetWithPermit\",\"inputs\":[{\"name\":\"bettor\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"faction\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"roundFactionPools\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"rounds\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"winningFaction\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"isSettled\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"totalPool\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"winnerPool\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"settledBlockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"settledBlockHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setMaxBetAmount\",\"inputs\":[{\"name\":\"_amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"settleRound\",\"inputs\":[{\"name\":\"roundId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"trustedForwarder\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"withdrawFunds\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"BetPlaced\",\"inputs\":[{\"name\":\"roundId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"faction\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"uint8\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MaxBetAmountUpdated\",\"inputs\":[{\"name\":\"newAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RewardsClaimed\",\"inputs\":[{\"name\":\"roundId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoundSettled\",\"inputs\":[{\"name\":\"roundId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"winningFaction\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"uint8\"},{\"name\":\"totalPool\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"winnerPool\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"settledBlockNumber\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"settledBlockHash\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"factionPools\",\"type\":\"uint256[4]\",\"indexed\":false,\"internalType\":\"uint256[4]\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AlreadyBet\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AlreadyClaimed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"BetAmountExceedsMax\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"BetAmountZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientContractBalance\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidFaction\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NoWinnerPool\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotWinner\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"RoundNotSettled\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"RoundNotStarted\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"RoundTimeLocked\",\"inputs\":[]}]",
	Bin: "0x60c060405234801561000f575f5ffd5b50604051612a99380380612a9983398181016040528101906100319190610316565b80335f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16036100a3575f6040517f1e4fbdf700000000000000000000000000000000000000000000000000000000815260040161009a9190610363565b60405180910390fd5b6100b2816101c560201b60201c565b5060016100d16100c661028660201b60201c565b6102af60201b60201c565b5f01819055508073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff1681525050505f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361017a576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610171906103d6565b60405180910390fd5b8173ffffffffffffffffffffffffffffffffffffffff1660a08173ffffffffffffffffffffffffffffffffffffffff1681525050683635c9adc5dea0000060018190555050506103f4565b5f5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050815f5f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b5f7f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f005f1b905090565b5f819050919050565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6102e5826102bc565b9050919050565b6102f5816102db565b81146102ff575f5ffd5b50565b5f81519050610310816102ec565b92915050565b5f5f6040838503121561032c5761032b6102b8565b5b5f61033985828601610302565b925050602061034a85828601610302565b9150509250929050565b61035d816102db565b82525050565b5f6020820190506103765f830184610354565b92915050565b5f82825260208201905092915050565b7f496e76616c6964207061796d656e7420746f6b656e00000000000000000000005f82015250565b5f6103c060158361037c565b91506103cb8261038c565b602082019050919050565b5f6020820190508181035f8301526103ed816103b4565b9050919050565b60805160a0516126686104315f395f818161045a015281816107e3015281816110370152818161110801526114d601525f61063f01526126685ff3fe608060405234801561000f575f5ffd5b506004361061014b575f3560e01c80638da5cb5b116100c1578063d2b48cd31161007a578063d2b48cd314610372578063dcb3b30e1461038e578063e3773fa5146103aa578063e9f0bff9146103da578063f2fde38b1461040a578063f644b3bb146104265761014b565b80638da5cb5b146102a8578063a525ad3c146102c6578063b428b2b5146102e4578063c107532914610300578063c642e7bf1461031c578063cab11d5d146103545761014b565b80636641ea08116101135780636641ea08146101f5578063715018a6146102135780637cfbc7a51461021d5780637da0a87714610239578063847f1907146102575780638c65c81f146102735761014b565b80633013ce291461014f578063350722ce1461016d5780635727e25d1461018b578063572b6c05146101a95780635eac6239146101d9575b5f5ffd5b610157610458565b6040516101649190611857565b60405180910390f35b61017561047c565b6040516101829190611888565b60405180910390f35b610193610481565b6040516101a09190611888565b60405180910390f35b6101c360048036038101906101be91906118e4565b6104a0565b6040516101d09190611929565b60405180910390f35b6101f360048036038101906101ee91906119a3565b6104de565b005b6101fd6105da565b60405161020a9190611888565b60405180910390f35b61021b6105e0565b005b61023760048036038101906102329190611a18565b6105f3565b005b61024161063c565b60405161024e9190611a52565b60405180910390f35b610271600480360381019061026c9190611a6b565b610663565b005b61028d60048036038101906102889190611a18565b61075c565b60405161029f96959493929190611afb565b60405180910390f35b6102b06107ac565b6040516102bd9190611a52565b60405180910390f35b6102ce6107d3565b6040516102db9190611888565b60405180910390f35b6102fe60048036038101906102f99190611bae565b6107d9565b005b61031a60048036038101906103159190611c4b565b6108c9565b005b61033660048036038101906103319190611a18565b610994565b60405161034b99989796959493929190611d2e565b60405180910390f35b61035c610a8d565b6040516103699190611888565b60405180910390f35b61038c60048036038101906103879190611a18565b610a93565b005b6103a860048036038101906103a39190611dbc565b610d45565b005b6103c460048036038101906103bf9190611a18565b610d6b565b6040516103d19190611929565b60405180910390f35b6103f460048036038101906103ef9190611dfa565b610d96565b6040516104019190611888565b60405180910390f35b610424600480360381019061041f91906118e4565b610dba565b005b610440600480360381019061043b9190611e38565b610e3e565b60405161044f93929190611e76565b60405180910390f35b7f000000000000000000000000000000000000000000000000000000000000000081565b607881565b5f610e1080426104919190611f05565b61049b9190611f35565b905090565b5f6104a961063c565b73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16149050919050565b6104e6610e88565b5f6104ef610eaa565b90505f5f90505f5f90505b848490508110156105b8575f6105298487878581811061051d5761051c611f76565b5b90506020020135610eb8565b90505f8111156105aa57808361053f9190611fa3565b92508373ffffffffffffffffffffffffffffffffffffffff1686868481811061056b5761056a611f76565b5b905060200201357f3300bdb359cfb956935bca32e9db727413eab1ca84341f2e36caea85bb796968836040516105a19190611888565b60405180910390a35b5080806001019150506104fa565b505f8111156105cc576105cb8282611034565b5b50506105d66111e8565b5050565b610e1081565b6105e8611202565b6105f15f611289565b565b6105fb611202565b806001819055507f0e858e4fc13b98505f129e4c64b83a3fcdfcfba2bd6b17d1a218390d19d0e769816040516106319190611888565b60405180910390a150565b5f7f0000000000000000000000000000000000000000000000000000000000000000905090565b61066b611202565b610673610e88565b5f5f90505f5f90505b8383905081101561073a575f6106ab8686868581811061069f5761069e611f76565b5b90506020020135610eb8565b90505f81111561072c5780836106c19190611fa3565b92508573ffffffffffffffffffffffffffffffffffffffff168585848181106106ed576106ec611f76565b5b905060200201357f3300bdb359cfb956935bca32e9db727413eab1ca84341f2e36caea85bb796968836040516107239190611888565b60405180910390a35b50808060010191505061067c565b505f81111561074e5761074d8482611034565b5b506107576111e8565b505050565b6002602052805f5260405f205f91509050805f015f9054906101000a900460ff1690805f0160019054906101000a900460ff16908060010154908060020154908060030154908060040154905086565b5f5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6101f481565b6107e1610e88565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663d505accf883088888888886040518863ffffffff1660e01b81526004016108469796959493929190611fd6565b5f604051808303815f87803b15801561085d575f5ffd5b505af192505050801561086e575060015b6108ad576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016108a49061209d565b60405180910390fd5b6108b887878761134a565b6108c06111e8565b50505050505050565b6108d1611202565b5f8273ffffffffffffffffffffffffffffffffffffffff1663a9059cbb33846040518363ffffffff1660e01b815260040161090d9291906120bb565b6020604051808303815f875af1158015610929573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061094d919061210c565b90508061098f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161098690612181565b60405180910390fd5b505050565b5f5f61099e6117bb565b5f5f5f5f5f5f5f8a036109ba576109b3610481565b98506109be565b8998505b5f60025f8b81526020019081526020015f2090508060010154985060045f8b81526020019081526020015f20600480602002604051908101604052809291908260048015610a21576020028201915b815481526020019060010190808311610a0d575b50505050509750610a3142610d6b565b8015610a435750610a40610481565b8a145b9650805f0160019054906101000a900460ff169550805f015f9054906101000a900460ff169450806002015493508060030154925080600401549150509193959799909294969850565b60015481565b610a9b610e88565b5f610e1082610aaa919061219f565b14610aea576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610ae190612219565b60405180910390fd5b610e1081610af89190611fa3565b421015610b31576040517f8e9c6e1c00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5f60025f8381526020019081526020015f209050805f0160019054906101000a900460ff1615610b96576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b8d90612281565b60405180910390fd5b5f816001015403610c27576001815f0160016101000a81548160ff021916908315150217905550817fc2b8ff06bd5d025468d7b1f59b34f98ac843be81a0cb25edcb0541f4e937add55f5f5f5f5f5f1b60405180608001604052805f81526020015f81526020015f81526020015f815250604051610c1996959493929190612308565b60405180910390a250610d3a565b5f600143610c359190612368565b4090505f815f1c90505f600482610c4c919061219f565b905080845f015f6101000a81548160ff021916908360ff1602179055506001845f0160016101000a81548160ff02191690831515021790555060045f8681526020019081526020015f208160ff1660048110610cab57610caa611f76565b5b01548460020181905550600143610cc29190612368565b8460030181905550828460040181905550847fc2b8ff06bd5d025468d7b1f59b34f98ac843be81a0cb25edcb0541f4e937add582866001015487600201548860030154896004015460045f8d81526020019081526020015f20604051610d2d96959493929190612456565b60405180910390a2505050505b610d426111e8565b50565b610d4d610e88565b610d5f610d58610eaa565b838361134a565b610d676111e8565b5050565b5f5f610e1083610d7b919061219f565b90506078610e10610d8c9190612368565b8110915050919050565b6004602052815f5260405f208160048110610daf575f80fd5b015f91509150505481565b610dc2611202565b5f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603610e32575f6040517f1e4fbdf7000000000000000000000000000000000000000000000000000000008152600401610e299190611a52565b60405180910390fd5b610e3b81611289565b50565b6003602052815f5260405f20602052805f5260405f205f9150915050805f015490806001015f9054906101000a900460ff16908060010160019054906101000a900460ff16905083565b610e90611665565b6002610ea2610e9d6116a6565b6116cf565b5f0181905550565b5f610eb36116d8565b905090565b5f5f60035f8481526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2090505f815f015403610f1b575f91505061102e565b8060010160019054906101000a900460ff1615610f3b575f91505061102e565b5f60025f8581526020019081526020015f209050805f0160019054906101000a900460ff16610f6e575f9250505061102e565b805f015f9054906101000a900460ff1660ff16826001015f9054906101000a900460ff1660ff1614610fa4575f9250505061102e565b5f816002015403610fb9575f9250505061102e565b5f81600201548260010154845f0154610fd29190611f35565b610fdc9190611f05565b90505f6127106101f4612710610ff29190612368565b83610ffd9190611f35565b6110079190611f05565b905060018460010160016101000a81548160ff021916908315150217905550809450505050505b92915050565b807f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166370a08231306040518263ffffffff1660e01b815260040161108e9190611a52565b602060405180830381865afa1580156110a9573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906110cd91906124ca565b1015611105576040517f786e0a9900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663a9059cbb84846040518363ffffffff1660e01b81526004016111619291906120bb565b6020604051808303815f875af115801561117d573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906111a1919061210c565b9050806111e3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016111da90612181565b60405180910390fd5b505050565b60016111fa6111f56116a6565b6116cf565b5f0181905550565b61120a610eaa565b73ffffffffffffffffffffffffffffffffffffffff166112286107ac565b73ffffffffffffffffffffffffffffffffffffffff16146112875761124b610eaa565b6040517f118cdaa700000000000000000000000000000000000000000000000000000000815260040161127e9190611a52565b60405180910390fd5b565b5f5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050815f5f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b60038260ff161115611388576040517fc8230a8100000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5f81036113c1576040517f61a725b400000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6001548111156113fd576040517f5bf2902500000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5f611406610481565b905061141142610d6b565b611447576040517f3f21028a00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5f60035f8381526020019081526020015f205f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2090505f815f015411156114d3576040517fa820742e00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166323b872dd8730876040518463ffffffff1660e01b8152600401611531939291906124f5565b6020604051808303815f875af115801561154d573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190611571919061210c565b9050806115b3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016115aa90612181565b60405180910390fd5b84826001015f6101000a81548160ff021916908360ff16021790555083825f01819055508360025f8581526020019081526020015f206001015f8282546115fa9190611fa3565b9250508190555061160c838686611740565b8573ffffffffffffffffffffffffffffffffffffffff16837f4f1eed5e863a822b0f9eb960dfdab2cc5a99beec4b191f2a7a9c7e28e5a15524878760405161165592919061252a565b60405180910390a3505050505050565b61166d611782565b156116a4576040517f3ee5aeb500000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b565b5f7f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f005f1b905090565b5f819050919050565b5f5f5f36905090505f6116e961179e565b905080821015801561170057506116ff336104a0565b5b15611730575f3682840390809261171993929190612559565b9061172491906125d4565b60601c9250505061173d565b6117386117ac565b925050505b90565b8060045f8581526020019081526020015f208360ff166004811061176757611766611f76565b5b015f8282546117769190611fa3565b92505081905550505050565b5f60026117956117906116a6565b6116cf565b5f015414905090565b5f6117a76117b3565b905090565b5f33905090565b5f6014905090565b6040518060800160405280600490602082028036833780820191505090505090565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f819050919050565b5f61181f61181a611815846117dd565b6117fc565b6117dd565b9050919050565b5f61183082611805565b9050919050565b5f61184182611826565b9050919050565b61185181611837565b82525050565b5f60208201905061186a5f830184611848565b92915050565b5f819050919050565b61188281611870565b82525050565b5f60208201905061189b5f830184611879565b92915050565b5f5ffd5b5f5ffd5b5f6118b3826117dd565b9050919050565b6118c3816118a9565b81146118cd575f5ffd5b50565b5f813590506118de816118ba565b92915050565b5f602082840312156118f9576118f86118a1565b5b5f611906848285016118d0565b91505092915050565b5f8115159050919050565b6119238161190f565b82525050565b5f60208201905061193c5f83018461191a565b92915050565b5f5ffd5b5f5ffd5b5f5ffd5b5f5f83601f84011261196357611962611942565b5b8235905067ffffffffffffffff8111156119805761197f611946565b5b60208301915083602082028301111561199c5761199b61194a565b5b9250929050565b5f5f602083850312156119b9576119b86118a1565b5b5f83013567ffffffffffffffff8111156119d6576119d56118a5565b5b6119e28582860161194e565b92509250509250929050565b6119f781611870565b8114611a01575f5ffd5b50565b5f81359050611a12816119ee565b92915050565b5f60208284031215611a2d57611a2c6118a1565b5b5f611a3a84828501611a04565b91505092915050565b611a4c816118a9565b82525050565b5f602082019050611a655f830184611a43565b92915050565b5f5f5f60408486031215611a8257611a816118a1565b5b5f611a8f868287016118d0565b935050602084013567ffffffffffffffff811115611ab057611aaf6118a5565b5b611abc8682870161194e565b92509250509250925092565b5f60ff82169050919050565b611add81611ac8565b82525050565b5f819050919050565b611af581611ae3565b82525050565b5f60c082019050611b0e5f830189611ad4565b611b1b602083018861191a565b611b286040830187611879565b611b356060830186611879565b611b426080830185611879565b611b4f60a0830184611aec565b979650505050505050565b611b6381611ac8565b8114611b6d575f5ffd5b50565b5f81359050611b7e81611b5a565b92915050565b611b8d81611ae3565b8114611b97575f5ffd5b50565b5f81359050611ba881611b84565b92915050565b5f5f5f5f5f5f5f60e0888a031215611bc957611bc86118a1565b5b5f611bd68a828b016118d0565b9750506020611be78a828b01611b70565b9650506040611bf88a828b01611a04565b9550506060611c098a828b01611a04565b9450506080611c1a8a828b01611b70565b93505060a0611c2b8a828b01611b9a565b92505060c0611c3c8a828b01611b9a565b91505092959891949750929550565b5f5f60408385031215611c6157611c606118a1565b5b5f611c6e858286016118d0565b9250506020611c7f85828601611a04565b9150509250929050565b5f60049050919050565b5f81905092915050565b5f819050919050565b611caf81611870565b82525050565b5f611cc08383611ca6565b60208301905092915050565b5f602082019050919050565b611ce181611c89565b611ceb8184611c93565b9250611cf682611c9d565b805f5b83811015611d26578151611d0d8782611cb5565b9650611d1883611ccc565b925050600181019050611cf9565b505050505050565b5f61018082019050611d425f83018c611879565b611d4f602083018b611879565b611d5c604083018a611cd8565b611d6960c083018961191a565b611d7660e083018861191a565b611d84610100830187611ad4565b611d92610120830186611879565b611da0610140830185611879565b611dae610160830184611aec565b9a9950505050505050505050565b5f5f60408385031215611dd257611dd16118a1565b5b5f611ddf85828601611b70565b9250506020611df085828601611a04565b9150509250929050565b5f5f60408385031215611e1057611e0f6118a1565b5b5f611e1d85828601611a04565b9250506020611e2e85828601611a04565b9150509250929050565b5f5f60408385031215611e4e57611e4d6118a1565b5b5f611e5b85828601611a04565b9250506020611e6c858286016118d0565b9150509250929050565b5f606082019050611e895f830186611879565b611e966020830185611ad4565b611ea3604083018461191a565b949350505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601260045260245ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f611f0f82611870565b9150611f1a83611870565b925082611f2a57611f29611eab565b5b828204905092915050565b5f611f3f82611870565b9150611f4a83611870565b9250828202611f5881611870565b91508282048414831517611f6f57611f6e611ed8565b5b5092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b5f611fad82611870565b9150611fb883611870565b9250828201905080821115611fd057611fcf611ed8565b5b92915050565b5f60e082019050611fe95f83018a611a43565b611ff66020830189611a43565b6120036040830188611879565b6120106060830187611879565b61201d6080830186611ad4565b61202a60a0830185611aec565b61203760c0830184611aec565b98975050505050505050565b5f82825260208201905092915050565b7f5065726d6974206661696c6564000000000000000000000000000000000000005f82015250565b5f612087600d83612043565b915061209282612053565b602082019050919050565b5f6020820190508181035f8301526120b48161207b565b9050919050565b5f6040820190506120ce5f830185611a43565b6120db6020830184611879565b9392505050565b6120eb8161190f565b81146120f5575f5ffd5b50565b5f81519050612106816120e2565b92915050565b5f60208284031215612121576121206118a1565b5b5f61212e848285016120f8565b91505092915050565b7f5472616e73666572206661696c656400000000000000000000000000000000005f82015250565b5f61216b600f83612043565b915061217682612137565b602082019050919050565b5f6020820190508181035f8301526121988161215f565b9050919050565b5f6121a982611870565b91506121b483611870565b9250826121c4576121c3611eab565b5b828206905092915050565b7f496e76616c696420726f756e64204944000000000000000000000000000000005f82015250565b5f612203601083612043565b915061220e826121cf565b602082019050919050565b5f6020820190508181035f830152612230816121f7565b9050919050565b7f416c726561647920736574746c656400000000000000000000000000000000005f82015250565b5f61226b600f83612043565b915061227682612237565b602082019050919050565b5f6020820190508181035f8301526122988161225f565b9050919050565b5f819050919050565b5f6122c26122bd6122b88461229f565b6117fc565b611ac8565b9050919050565b6122d2816122a8565b82525050565b5f6122f26122ed6122e88461229f565b6117fc565b611870565b9050919050565b612302816122d8565b82525050565b5f6101208201905061231c5f8301896122c9565b61232960208301886122f9565b61233660408301876122f9565b61234360608301866122f9565b6123506080830185611aec565b61235d60a0830184611cd8565b979650505050505050565b5f61237282611870565b915061237d83611870565b925082820390508181111561239557612394611ed8565b5b92915050565b5f60049050919050565b5f819050919050565b5f815f1c9050919050565b5f819050919050565b5f6123d46123cf836123ae565b6123b9565b9050919050565b5f6123e682546123c2565b9050919050565b5f600182019050919050565b6124028161239b565b61240c8184611c93565b9250612417826123a5565b805f5b8381101561244e5761242b826123db565b6124358782611cb5565b9650612440836123ed565b92505060018101905061241a565b505050505050565b5f6101208201905061246a5f830189611ad4565b6124776020830188611879565b6124846040830187611879565b6124916060830186611879565b61249e6080830185611aec565b6124ab60a08301846123f9565b979650505050505050565b5f815190506124c4816119ee565b92915050565b5f602082840312156124df576124de6118a1565b5b5f6124ec848285016124b6565b91505092915050565b5f6060820190506125085f830186611a43565b6125156020830185611a43565b6125226040830184611879565b949350505050565b5f60408201905061253d5f830185611ad4565b61254a6020830184611879565b9392505050565b5f5ffd5b5f5ffd5b5f5f8585111561256c5761256b612551565b5b8386111561257d5761257c612555565b5b6001850283019150848603905094509492505050565b5f82905092915050565b5f7fffffffffffffffffffffffffffffffffffffffff00000000000000000000000082169050919050565b5f82821b905092915050565b5f6125df8383612593565b826125ea813561259d565b9250601482101561262a576126257fffffffffffffffffffffffffffffffffffffffff000000000000000000000000836014036008026125c8565b831692505b50509291505056fea264697066735822122002fe0c5c14dec2c5939d12a8eb175ff892f90192e33998fb7be4c6c752cc9f3964736f6c63430008210033",
}

// HourlySquadGameABI is the input ABI used to generate the binding from.
// Deprecated: Use HourlySquadGameMetaData.ABI instead.
var HourlySquadGameABI = HourlySquadGameMetaData.ABI

// HourlySquadGameBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use HourlySquadGameMetaData.Bin instead.
var HourlySquadGameBin = HourlySquadGameMetaData.Bin

// DeployHourlySquadGame deploys a new Ethereum contract, binding an instance of HourlySquadGame to it.
func DeployHourlySquadGame(auth *bind.TransactOpts, backend bind.ContractBackend, _paymentToken common.Address, _trustedForwarder common.Address) (common.Address, *types.Transaction, *HourlySquadGame, error) {
	parsed, err := HourlySquadGameMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(HourlySquadGameBin), backend, _paymentToken, _trustedForwarder)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &HourlySquadGame{HourlySquadGameCaller: HourlySquadGameCaller{contract: contract}, HourlySquadGameTransactor: HourlySquadGameTransactor{contract: contract}, HourlySquadGameFilterer: HourlySquadGameFilterer{contract: contract}}, nil
}

// HourlySquadGame is an auto generated Go binding around an Ethereum contract.
type HourlySquadGame struct {
	HourlySquadGameCaller     // Read-only binding to the contract
	HourlySquadGameTransactor // Write-only binding to the contract
	HourlySquadGameFilterer   // Log filterer for contract events
}

// HourlySquadGameCaller is an auto generated read-only Go binding around an Ethereum contract.
type HourlySquadGameCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HourlySquadGameTransactor is an auto generated write-only Go binding around an Ethereum contract.
type HourlySquadGameTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HourlySquadGameFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type HourlySquadGameFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HourlySquadGameSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HourlySquadGameSession struct {
	Contract     *HourlySquadGame  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HourlySquadGameCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HourlySquadGameCallerSession struct {
	Contract *HourlySquadGameCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// HourlySquadGameTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HourlySquadGameTransactorSession struct {
	Contract     *HourlySquadGameTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// HourlySquadGameRaw is an auto generated low-level Go binding around an Ethereum contract.
type HourlySquadGameRaw struct {
	Contract *HourlySquadGame // Generic contract binding to access the raw methods on
}

// HourlySquadGameCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HourlySquadGameCallerRaw struct {
	Contract *HourlySquadGameCaller // Generic read-only contract binding to access the raw methods on
}

// HourlySquadGameTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HourlySquadGameTransactorRaw struct {
	Contract *HourlySquadGameTransactor // Generic write-only contract binding to access the raw methods on
}

// NewHourlySquadGame creates a new instance of HourlySquadGame, bound to a specific deployed contract.
func NewHourlySquadGame(address common.Address, backend bind.ContractBackend) (*HourlySquadGame, error) {
	contract, err := bindHourlySquadGame(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &HourlySquadGame{HourlySquadGameCaller: HourlySquadGameCaller{contract: contract}, HourlySquadGameTransactor: HourlySquadGameTransactor{contract: contract}, HourlySquadGameFilterer: HourlySquadGameFilterer{contract: contract}}, nil
}

// NewHourlySquadGameCaller creates a new read-only instance of HourlySquadGame, bound to a specific deployed contract.
func NewHourlySquadGameCaller(address common.Address, caller bind.ContractCaller) (*HourlySquadGameCaller, error) {
	contract, err := bindHourlySquadGame(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &HourlySquadGameCaller{contract: contract}, nil
}

// NewHourlySquadGameTransactor creates a new write-only instance of HourlySquadGame, bound to a specific deployed contract.
func NewHourlySquadGameTransactor(address common.Address, transactor bind.ContractTransactor) (*HourlySquadGameTransactor, error) {
	contract, err := bindHourlySquadGame(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &HourlySquadGameTransactor{contract: contract}, nil
}

// NewHourlySquadGameFilterer creates a new log filterer instance of HourlySquadGame, bound to a specific deployed contract.
func NewHourlySquadGameFilterer(address common.Address, filterer bind.ContractFilterer) (*HourlySquadGameFilterer, error) {
	contract, err := bindHourlySquadGame(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &HourlySquadGameFilterer{contract: contract}, nil
}

// bindHourlySquadGame binds a generic wrapper to an already deployed contract.
func bindHourlySquadGame(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := HourlySquadGameMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HourlySquadGame *HourlySquadGameRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HourlySquadGame.Contract.HourlySquadGameCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HourlySquadGame *HourlySquadGameRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.HourlySquadGameTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HourlySquadGame *HourlySquadGameRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.HourlySquadGameTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HourlySquadGame *HourlySquadGameCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HourlySquadGame.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HourlySquadGame *HourlySquadGameTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HourlySquadGame *HourlySquadGameTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.contract.Transact(opts, method, params...)
}

// FEEBASISPOINTS is a free data retrieval call binding the contract method 0xa525ad3c.
//
// Solidity: function FEE_BASIS_POINTS() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameCaller) FEEBASISPOINTS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "FEE_BASIS_POINTS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FEEBASISPOINTS is a free data retrieval call binding the contract method 0xa525ad3c.
//
// Solidity: function FEE_BASIS_POINTS() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameSession) FEEBASISPOINTS() (*big.Int, error) {
	return _HourlySquadGame.Contract.FEEBASISPOINTS(&_HourlySquadGame.CallOpts)
}

// FEEBASISPOINTS is a free data retrieval call binding the contract method 0xa525ad3c.
//
// Solidity: function FEE_BASIS_POINTS() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameCallerSession) FEEBASISPOINTS() (*big.Int, error) {
	return _HourlySquadGame.Contract.FEEBASISPOINTS(&_HourlySquadGame.CallOpts)
}

// LOCKWINDOW is a free data retrieval call binding the contract method 0x350722ce.
//
// Solidity: function LOCK_WINDOW() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameCaller) LOCKWINDOW(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "LOCK_WINDOW")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LOCKWINDOW is a free data retrieval call binding the contract method 0x350722ce.
//
// Solidity: function LOCK_WINDOW() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameSession) LOCKWINDOW() (*big.Int, error) {
	return _HourlySquadGame.Contract.LOCKWINDOW(&_HourlySquadGame.CallOpts)
}

// LOCKWINDOW is a free data retrieval call binding the contract method 0x350722ce.
//
// Solidity: function LOCK_WINDOW() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameCallerSession) LOCKWINDOW() (*big.Int, error) {
	return _HourlySquadGame.Contract.LOCKWINDOW(&_HourlySquadGame.CallOpts)
}

// ROUNDDURATION is a free data retrieval call binding the contract method 0x6641ea08.
//
// Solidity: function ROUND_DURATION() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameCaller) ROUNDDURATION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "ROUND_DURATION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ROUNDDURATION is a free data retrieval call binding the contract method 0x6641ea08.
//
// Solidity: function ROUND_DURATION() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameSession) ROUNDDURATION() (*big.Int, error) {
	return _HourlySquadGame.Contract.ROUNDDURATION(&_HourlySquadGame.CallOpts)
}

// ROUNDDURATION is a free data retrieval call binding the contract method 0x6641ea08.
//
// Solidity: function ROUND_DURATION() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameCallerSession) ROUNDDURATION() (*big.Int, error) {
	return _HourlySquadGame.Contract.ROUNDDURATION(&_HourlySquadGame.CallOpts)
}

// Bets is a free data retrieval call binding the contract method 0xf644b3bb.
//
// Solidity: function bets(uint256 , address ) view returns(uint256 amount, uint8 faction, bool isClaimed)
func (_HourlySquadGame *HourlySquadGameCaller) Bets(opts *bind.CallOpts, arg0 *big.Int, arg1 common.Address) (struct {
	Amount    *big.Int
	Faction   uint8
	IsClaimed bool
}, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "bets", arg0, arg1)

	outstruct := new(struct {
		Amount    *big.Int
		Faction   uint8
		IsClaimed bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Amount = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Faction = *abi.ConvertType(out[1], new(uint8)).(*uint8)
	outstruct.IsClaimed = *abi.ConvertType(out[2], new(bool)).(*bool)

	return *outstruct, err

}

// Bets is a free data retrieval call binding the contract method 0xf644b3bb.
//
// Solidity: function bets(uint256 , address ) view returns(uint256 amount, uint8 faction, bool isClaimed)
func (_HourlySquadGame *HourlySquadGameSession) Bets(arg0 *big.Int, arg1 common.Address) (struct {
	Amount    *big.Int
	Faction   uint8
	IsClaimed bool
}, error) {
	return _HourlySquadGame.Contract.Bets(&_HourlySquadGame.CallOpts, arg0, arg1)
}

// Bets is a free data retrieval call binding the contract method 0xf644b3bb.
//
// Solidity: function bets(uint256 , address ) view returns(uint256 amount, uint8 faction, bool isClaimed)
func (_HourlySquadGame *HourlySquadGameCallerSession) Bets(arg0 *big.Int, arg1 common.Address) (struct {
	Amount    *big.Int
	Faction   uint8
	IsClaimed bool
}, error) {
	return _HourlySquadGame.Contract.Bets(&_HourlySquadGame.CallOpts, arg0, arg1)
}

// GetCurrentRoundId is a free data retrieval call binding the contract method 0x5727e25d.
//
// Solidity: function getCurrentRoundId() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameCaller) GetCurrentRoundId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "getCurrentRoundId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCurrentRoundId is a free data retrieval call binding the contract method 0x5727e25d.
//
// Solidity: function getCurrentRoundId() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameSession) GetCurrentRoundId() (*big.Int, error) {
	return _HourlySquadGame.Contract.GetCurrentRoundId(&_HourlySquadGame.CallOpts)
}

// GetCurrentRoundId is a free data retrieval call binding the contract method 0x5727e25d.
//
// Solidity: function getCurrentRoundId() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameCallerSession) GetCurrentRoundId() (*big.Int, error) {
	return _HourlySquadGame.Contract.GetCurrentRoundId(&_HourlySquadGame.CallOpts)
}

// GetRoundState is a free data retrieval call binding the contract method 0xc642e7bf.
//
// Solidity: function getRoundState(uint256 _roundId) view returns(uint256 roundId, uint256 totalPool, uint256[4] factionPools, bool isOpen, bool isSettled, uint8 winningFaction, uint256 winnerPool, uint256 settledBlockNumber, bytes32 settledBlockHash)
func (_HourlySquadGame *HourlySquadGameCaller) GetRoundState(opts *bind.CallOpts, _roundId *big.Int) (struct {
	RoundId            *big.Int
	TotalPool          *big.Int
	FactionPools       [4]*big.Int
	IsOpen             bool
	IsSettled          bool
	WinningFaction     uint8
	WinnerPool         *big.Int
	SettledBlockNumber *big.Int
	SettledBlockHash   [32]byte
}, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "getRoundState", _roundId)

	outstruct := new(struct {
		RoundId            *big.Int
		TotalPool          *big.Int
		FactionPools       [4]*big.Int
		IsOpen             bool
		IsSettled          bool
		WinningFaction     uint8
		WinnerPool         *big.Int
		SettledBlockNumber *big.Int
		SettledBlockHash   [32]byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.RoundId = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.TotalPool = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.FactionPools = *abi.ConvertType(out[2], new([4]*big.Int)).(*[4]*big.Int)
	outstruct.IsOpen = *abi.ConvertType(out[3], new(bool)).(*bool)
	outstruct.IsSettled = *abi.ConvertType(out[4], new(bool)).(*bool)
	outstruct.WinningFaction = *abi.ConvertType(out[5], new(uint8)).(*uint8)
	outstruct.WinnerPool = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.SettledBlockNumber = *abi.ConvertType(out[7], new(*big.Int)).(**big.Int)
	outstruct.SettledBlockHash = *abi.ConvertType(out[8], new([32]byte)).(*[32]byte)

	return *outstruct, err

}

// GetRoundState is a free data retrieval call binding the contract method 0xc642e7bf.
//
// Solidity: function getRoundState(uint256 _roundId) view returns(uint256 roundId, uint256 totalPool, uint256[4] factionPools, bool isOpen, bool isSettled, uint8 winningFaction, uint256 winnerPool, uint256 settledBlockNumber, bytes32 settledBlockHash)
func (_HourlySquadGame *HourlySquadGameSession) GetRoundState(_roundId *big.Int) (struct {
	RoundId            *big.Int
	TotalPool          *big.Int
	FactionPools       [4]*big.Int
	IsOpen             bool
	IsSettled          bool
	WinningFaction     uint8
	WinnerPool         *big.Int
	SettledBlockNumber *big.Int
	SettledBlockHash   [32]byte
}, error) {
	return _HourlySquadGame.Contract.GetRoundState(&_HourlySquadGame.CallOpts, _roundId)
}

// GetRoundState is a free data retrieval call binding the contract method 0xc642e7bf.
//
// Solidity: function getRoundState(uint256 _roundId) view returns(uint256 roundId, uint256 totalPool, uint256[4] factionPools, bool isOpen, bool isSettled, uint8 winningFaction, uint256 winnerPool, uint256 settledBlockNumber, bytes32 settledBlockHash)
func (_HourlySquadGame *HourlySquadGameCallerSession) GetRoundState(_roundId *big.Int) (struct {
	RoundId            *big.Int
	TotalPool          *big.Int
	FactionPools       [4]*big.Int
	IsOpen             bool
	IsSettled          bool
	WinningFaction     uint8
	WinnerPool         *big.Int
	SettledBlockNumber *big.Int
	SettledBlockHash   [32]byte
}, error) {
	return _HourlySquadGame.Contract.GetRoundState(&_HourlySquadGame.CallOpts, _roundId)
}

// IsBettingOpen is a free data retrieval call binding the contract method 0xe3773fa5.
//
// Solidity: function isBettingOpen(uint256 timestamp) pure returns(bool)
func (_HourlySquadGame *HourlySquadGameCaller) IsBettingOpen(opts *bind.CallOpts, timestamp *big.Int) (bool, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "isBettingOpen", timestamp)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsBettingOpen is a free data retrieval call binding the contract method 0xe3773fa5.
//
// Solidity: function isBettingOpen(uint256 timestamp) pure returns(bool)
func (_HourlySquadGame *HourlySquadGameSession) IsBettingOpen(timestamp *big.Int) (bool, error) {
	return _HourlySquadGame.Contract.IsBettingOpen(&_HourlySquadGame.CallOpts, timestamp)
}

// IsBettingOpen is a free data retrieval call binding the contract method 0xe3773fa5.
//
// Solidity: function isBettingOpen(uint256 timestamp) pure returns(bool)
func (_HourlySquadGame *HourlySquadGameCallerSession) IsBettingOpen(timestamp *big.Int) (bool, error) {
	return _HourlySquadGame.Contract.IsBettingOpen(&_HourlySquadGame.CallOpts, timestamp)
}

// IsTrustedForwarder is a free data retrieval call binding the contract method 0x572b6c05.
//
// Solidity: function isTrustedForwarder(address forwarder) view returns(bool)
func (_HourlySquadGame *HourlySquadGameCaller) IsTrustedForwarder(opts *bind.CallOpts, forwarder common.Address) (bool, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "isTrustedForwarder", forwarder)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsTrustedForwarder is a free data retrieval call binding the contract method 0x572b6c05.
//
// Solidity: function isTrustedForwarder(address forwarder) view returns(bool)
func (_HourlySquadGame *HourlySquadGameSession) IsTrustedForwarder(forwarder common.Address) (bool, error) {
	return _HourlySquadGame.Contract.IsTrustedForwarder(&_HourlySquadGame.CallOpts, forwarder)
}

// IsTrustedForwarder is a free data retrieval call binding the contract method 0x572b6c05.
//
// Solidity: function isTrustedForwarder(address forwarder) view returns(bool)
func (_HourlySquadGame *HourlySquadGameCallerSession) IsTrustedForwarder(forwarder common.Address) (bool, error) {
	return _HourlySquadGame.Contract.IsTrustedForwarder(&_HourlySquadGame.CallOpts, forwarder)
}

// MaxBetAmount is a free data retrieval call binding the contract method 0xcab11d5d.
//
// Solidity: function maxBetAmount() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameCaller) MaxBetAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "maxBetAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxBetAmount is a free data retrieval call binding the contract method 0xcab11d5d.
//
// Solidity: function maxBetAmount() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameSession) MaxBetAmount() (*big.Int, error) {
	return _HourlySquadGame.Contract.MaxBetAmount(&_HourlySquadGame.CallOpts)
}

// MaxBetAmount is a free data retrieval call binding the contract method 0xcab11d5d.
//
// Solidity: function maxBetAmount() view returns(uint256)
func (_HourlySquadGame *HourlySquadGameCallerSession) MaxBetAmount() (*big.Int, error) {
	return _HourlySquadGame.Contract.MaxBetAmount(&_HourlySquadGame.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_HourlySquadGame *HourlySquadGameCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_HourlySquadGame *HourlySquadGameSession) Owner() (common.Address, error) {
	return _HourlySquadGame.Contract.Owner(&_HourlySquadGame.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_HourlySquadGame *HourlySquadGameCallerSession) Owner() (common.Address, error) {
	return _HourlySquadGame.Contract.Owner(&_HourlySquadGame.CallOpts)
}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_HourlySquadGame *HourlySquadGameCaller) PaymentToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "paymentToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_HourlySquadGame *HourlySquadGameSession) PaymentToken() (common.Address, error) {
	return _HourlySquadGame.Contract.PaymentToken(&_HourlySquadGame.CallOpts)
}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_HourlySquadGame *HourlySquadGameCallerSession) PaymentToken() (common.Address, error) {
	return _HourlySquadGame.Contract.PaymentToken(&_HourlySquadGame.CallOpts)
}

// RoundFactionPools is a free data retrieval call binding the contract method 0xe9f0bff9.
//
// Solidity: function roundFactionPools(uint256 , uint256 ) view returns(uint256)
func (_HourlySquadGame *HourlySquadGameCaller) RoundFactionPools(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "roundFactionPools", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RoundFactionPools is a free data retrieval call binding the contract method 0xe9f0bff9.
//
// Solidity: function roundFactionPools(uint256 , uint256 ) view returns(uint256)
func (_HourlySquadGame *HourlySquadGameSession) RoundFactionPools(arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	return _HourlySquadGame.Contract.RoundFactionPools(&_HourlySquadGame.CallOpts, arg0, arg1)
}

// RoundFactionPools is a free data retrieval call binding the contract method 0xe9f0bff9.
//
// Solidity: function roundFactionPools(uint256 , uint256 ) view returns(uint256)
func (_HourlySquadGame *HourlySquadGameCallerSession) RoundFactionPools(arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	return _HourlySquadGame.Contract.RoundFactionPools(&_HourlySquadGame.CallOpts, arg0, arg1)
}

// Rounds is a free data retrieval call binding the contract method 0x8c65c81f.
//
// Solidity: function rounds(uint256 ) view returns(uint8 winningFaction, bool isSettled, uint256 totalPool, uint256 winnerPool, uint256 settledBlockNumber, bytes32 settledBlockHash)
func (_HourlySquadGame *HourlySquadGameCaller) Rounds(opts *bind.CallOpts, arg0 *big.Int) (struct {
	WinningFaction     uint8
	IsSettled          bool
	TotalPool          *big.Int
	WinnerPool         *big.Int
	SettledBlockNumber *big.Int
	SettledBlockHash   [32]byte
}, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "rounds", arg0)

	outstruct := new(struct {
		WinningFaction     uint8
		IsSettled          bool
		TotalPool          *big.Int
		WinnerPool         *big.Int
		SettledBlockNumber *big.Int
		SettledBlockHash   [32]byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.WinningFaction = *abi.ConvertType(out[0], new(uint8)).(*uint8)
	outstruct.IsSettled = *abi.ConvertType(out[1], new(bool)).(*bool)
	outstruct.TotalPool = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.WinnerPool = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.SettledBlockNumber = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.SettledBlockHash = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)

	return *outstruct, err

}

// Rounds is a free data retrieval call binding the contract method 0x8c65c81f.
//
// Solidity: function rounds(uint256 ) view returns(uint8 winningFaction, bool isSettled, uint256 totalPool, uint256 winnerPool, uint256 settledBlockNumber, bytes32 settledBlockHash)
func (_HourlySquadGame *HourlySquadGameSession) Rounds(arg0 *big.Int) (struct {
	WinningFaction     uint8
	IsSettled          bool
	TotalPool          *big.Int
	WinnerPool         *big.Int
	SettledBlockNumber *big.Int
	SettledBlockHash   [32]byte
}, error) {
	return _HourlySquadGame.Contract.Rounds(&_HourlySquadGame.CallOpts, arg0)
}

// Rounds is a free data retrieval call binding the contract method 0x8c65c81f.
//
// Solidity: function rounds(uint256 ) view returns(uint8 winningFaction, bool isSettled, uint256 totalPool, uint256 winnerPool, uint256 settledBlockNumber, bytes32 settledBlockHash)
func (_HourlySquadGame *HourlySquadGameCallerSession) Rounds(arg0 *big.Int) (struct {
	WinningFaction     uint8
	IsSettled          bool
	TotalPool          *big.Int
	WinnerPool         *big.Int
	SettledBlockNumber *big.Int
	SettledBlockHash   [32]byte
}, error) {
	return _HourlySquadGame.Contract.Rounds(&_HourlySquadGame.CallOpts, arg0)
}

// TrustedForwarder is a free data retrieval call binding the contract method 0x7da0a877.
//
// Solidity: function trustedForwarder() view returns(address)
func (_HourlySquadGame *HourlySquadGameCaller) TrustedForwarder(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HourlySquadGame.contract.Call(opts, &out, "trustedForwarder")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TrustedForwarder is a free data retrieval call binding the contract method 0x7da0a877.
//
// Solidity: function trustedForwarder() view returns(address)
func (_HourlySquadGame *HourlySquadGameSession) TrustedForwarder() (common.Address, error) {
	return _HourlySquadGame.Contract.TrustedForwarder(&_HourlySquadGame.CallOpts)
}

// TrustedForwarder is a free data retrieval call binding the contract method 0x7da0a877.
//
// Solidity: function trustedForwarder() view returns(address)
func (_HourlySquadGame *HourlySquadGameCallerSession) TrustedForwarder() (common.Address, error) {
	return _HourlySquadGame.Contract.TrustedForwarder(&_HourlySquadGame.CallOpts)
}

// ClaimRewards is a paid mutator transaction binding the contract method 0x5eac6239.
//
// Solidity: function claimRewards(uint256[] roundIds) returns()
func (_HourlySquadGame *HourlySquadGameTransactor) ClaimRewards(opts *bind.TransactOpts, roundIds []*big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.contract.Transact(opts, "claimRewards", roundIds)
}

// ClaimRewards is a paid mutator transaction binding the contract method 0x5eac6239.
//
// Solidity: function claimRewards(uint256[] roundIds) returns()
func (_HourlySquadGame *HourlySquadGameSession) ClaimRewards(roundIds []*big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.ClaimRewards(&_HourlySquadGame.TransactOpts, roundIds)
}

// ClaimRewards is a paid mutator transaction binding the contract method 0x5eac6239.
//
// Solidity: function claimRewards(uint256[] roundIds) returns()
func (_HourlySquadGame *HourlySquadGameTransactorSession) ClaimRewards(roundIds []*big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.ClaimRewards(&_HourlySquadGame.TransactOpts, roundIds)
}

// DistributeRewards is a paid mutator transaction binding the contract method 0x847f1907.
//
// Solidity: function distributeRewards(address user, uint256[] roundIds) returns()
func (_HourlySquadGame *HourlySquadGameTransactor) DistributeRewards(opts *bind.TransactOpts, user common.Address, roundIds []*big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.contract.Transact(opts, "distributeRewards", user, roundIds)
}

// DistributeRewards is a paid mutator transaction binding the contract method 0x847f1907.
//
// Solidity: function distributeRewards(address user, uint256[] roundIds) returns()
func (_HourlySquadGame *HourlySquadGameSession) DistributeRewards(user common.Address, roundIds []*big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.DistributeRewards(&_HourlySquadGame.TransactOpts, user, roundIds)
}

// DistributeRewards is a paid mutator transaction binding the contract method 0x847f1907.
//
// Solidity: function distributeRewards(address user, uint256[] roundIds) returns()
func (_HourlySquadGame *HourlySquadGameTransactorSession) DistributeRewards(user common.Address, roundIds []*big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.DistributeRewards(&_HourlySquadGame.TransactOpts, user, roundIds)
}

// PlaceBet is a paid mutator transaction binding the contract method 0xdcb3b30e.
//
// Solidity: function placeBet(uint8 faction, uint256 amount) returns()
func (_HourlySquadGame *HourlySquadGameTransactor) PlaceBet(opts *bind.TransactOpts, faction uint8, amount *big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.contract.Transact(opts, "placeBet", faction, amount)
}

// PlaceBet is a paid mutator transaction binding the contract method 0xdcb3b30e.
//
// Solidity: function placeBet(uint8 faction, uint256 amount) returns()
func (_HourlySquadGame *HourlySquadGameSession) PlaceBet(faction uint8, amount *big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.PlaceBet(&_HourlySquadGame.TransactOpts, faction, amount)
}

// PlaceBet is a paid mutator transaction binding the contract method 0xdcb3b30e.
//
// Solidity: function placeBet(uint8 faction, uint256 amount) returns()
func (_HourlySquadGame *HourlySquadGameTransactorSession) PlaceBet(faction uint8, amount *big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.PlaceBet(&_HourlySquadGame.TransactOpts, faction, amount)
}

// PlaceBetWithPermit is a paid mutator transaction binding the contract method 0xb428b2b5.
//
// Solidity: function placeBetWithPermit(address bettor, uint8 faction, uint256 amount, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_HourlySquadGame *HourlySquadGameTransactor) PlaceBetWithPermit(opts *bind.TransactOpts, bettor common.Address, faction uint8, amount *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _HourlySquadGame.contract.Transact(opts, "placeBetWithPermit", bettor, faction, amount, deadline, v, r, s)
}

// PlaceBetWithPermit is a paid mutator transaction binding the contract method 0xb428b2b5.
//
// Solidity: function placeBetWithPermit(address bettor, uint8 faction, uint256 amount, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_HourlySquadGame *HourlySquadGameSession) PlaceBetWithPermit(bettor common.Address, faction uint8, amount *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.PlaceBetWithPermit(&_HourlySquadGame.TransactOpts, bettor, faction, amount, deadline, v, r, s)
}

// PlaceBetWithPermit is a paid mutator transaction binding the contract method 0xb428b2b5.
//
// Solidity: function placeBetWithPermit(address bettor, uint8 faction, uint256 amount, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_HourlySquadGame *HourlySquadGameTransactorSession) PlaceBetWithPermit(bettor common.Address, faction uint8, amount *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.PlaceBetWithPermit(&_HourlySquadGame.TransactOpts, bettor, faction, amount, deadline, v, r, s)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_HourlySquadGame *HourlySquadGameTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HourlySquadGame.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_HourlySquadGame *HourlySquadGameSession) RenounceOwnership() (*types.Transaction, error) {
	return _HourlySquadGame.Contract.RenounceOwnership(&_HourlySquadGame.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_HourlySquadGame *HourlySquadGameTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _HourlySquadGame.Contract.RenounceOwnership(&_HourlySquadGame.TransactOpts)
}

// SetMaxBetAmount is a paid mutator transaction binding the contract method 0x7cfbc7a5.
//
// Solidity: function setMaxBetAmount(uint256 _amount) returns()
func (_HourlySquadGame *HourlySquadGameTransactor) SetMaxBetAmount(opts *bind.TransactOpts, _amount *big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.contract.Transact(opts, "setMaxBetAmount", _amount)
}

// SetMaxBetAmount is a paid mutator transaction binding the contract method 0x7cfbc7a5.
//
// Solidity: function setMaxBetAmount(uint256 _amount) returns()
func (_HourlySquadGame *HourlySquadGameSession) SetMaxBetAmount(_amount *big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.SetMaxBetAmount(&_HourlySquadGame.TransactOpts, _amount)
}

// SetMaxBetAmount is a paid mutator transaction binding the contract method 0x7cfbc7a5.
//
// Solidity: function setMaxBetAmount(uint256 _amount) returns()
func (_HourlySquadGame *HourlySquadGameTransactorSession) SetMaxBetAmount(_amount *big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.SetMaxBetAmount(&_HourlySquadGame.TransactOpts, _amount)
}

// SettleRound is a paid mutator transaction binding the contract method 0xd2b48cd3.
//
// Solidity: function settleRound(uint256 roundId) returns()
func (_HourlySquadGame *HourlySquadGameTransactor) SettleRound(opts *bind.TransactOpts, roundId *big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.contract.Transact(opts, "settleRound", roundId)
}

// SettleRound is a paid mutator transaction binding the contract method 0xd2b48cd3.
//
// Solidity: function settleRound(uint256 roundId) returns()
func (_HourlySquadGame *HourlySquadGameSession) SettleRound(roundId *big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.SettleRound(&_HourlySquadGame.TransactOpts, roundId)
}

// SettleRound is a paid mutator transaction binding the contract method 0xd2b48cd3.
//
// Solidity: function settleRound(uint256 roundId) returns()
func (_HourlySquadGame *HourlySquadGameTransactorSession) SettleRound(roundId *big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.SettleRound(&_HourlySquadGame.TransactOpts, roundId)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_HourlySquadGame *HourlySquadGameTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _HourlySquadGame.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_HourlySquadGame *HourlySquadGameSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.TransferOwnership(&_HourlySquadGame.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_HourlySquadGame *HourlySquadGameTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.TransferOwnership(&_HourlySquadGame.TransactOpts, newOwner)
}

// WithdrawFunds is a paid mutator transaction binding the contract method 0xc1075329.
//
// Solidity: function withdrawFunds(address token, uint256 amount) returns()
func (_HourlySquadGame *HourlySquadGameTransactor) WithdrawFunds(opts *bind.TransactOpts, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.contract.Transact(opts, "withdrawFunds", token, amount)
}

// WithdrawFunds is a paid mutator transaction binding the contract method 0xc1075329.
//
// Solidity: function withdrawFunds(address token, uint256 amount) returns()
func (_HourlySquadGame *HourlySquadGameSession) WithdrawFunds(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.WithdrawFunds(&_HourlySquadGame.TransactOpts, token, amount)
}

// WithdrawFunds is a paid mutator transaction binding the contract method 0xc1075329.
//
// Solidity: function withdrawFunds(address token, uint256 amount) returns()
func (_HourlySquadGame *HourlySquadGameTransactorSession) WithdrawFunds(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _HourlySquadGame.Contract.WithdrawFunds(&_HourlySquadGame.TransactOpts, token, amount)
}

// HourlySquadGameBetPlacedIterator is returned from FilterBetPlaced and is used to iterate over the raw logs and unpacked data for BetPlaced events raised by the HourlySquadGame contract.
type HourlySquadGameBetPlacedIterator struct {
	Event *HourlySquadGameBetPlaced // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *HourlySquadGameBetPlacedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HourlySquadGameBetPlaced)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(HourlySquadGameBetPlaced)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *HourlySquadGameBetPlacedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HourlySquadGameBetPlacedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HourlySquadGameBetPlaced represents a BetPlaced event raised by the HourlySquadGame contract.
type HourlySquadGameBetPlaced struct {
	RoundId *big.Int
	User    common.Address
	Faction uint8
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterBetPlaced is a free log retrieval operation binding the contract event 0x4f1eed5e863a822b0f9eb960dfdab2cc5a99beec4b191f2a7a9c7e28e5a15524.
//
// Solidity: event BetPlaced(uint256 indexed roundId, address indexed user, uint8 faction, uint256 amount)
func (_HourlySquadGame *HourlySquadGameFilterer) FilterBetPlaced(opts *bind.FilterOpts, roundId []*big.Int, user []common.Address) (*HourlySquadGameBetPlacedIterator, error) {

	var roundIdRule []interface{}
	for _, roundIdItem := range roundId {
		roundIdRule = append(roundIdRule, roundIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _HourlySquadGame.contract.FilterLogs(opts, "BetPlaced", roundIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return &HourlySquadGameBetPlacedIterator{contract: _HourlySquadGame.contract, event: "BetPlaced", logs: logs, sub: sub}, nil
}

// WatchBetPlaced is a free log subscription operation binding the contract event 0x4f1eed5e863a822b0f9eb960dfdab2cc5a99beec4b191f2a7a9c7e28e5a15524.
//
// Solidity: event BetPlaced(uint256 indexed roundId, address indexed user, uint8 faction, uint256 amount)
func (_HourlySquadGame *HourlySquadGameFilterer) WatchBetPlaced(opts *bind.WatchOpts, sink chan<- *HourlySquadGameBetPlaced, roundId []*big.Int, user []common.Address) (event.Subscription, error) {

	var roundIdRule []interface{}
	for _, roundIdItem := range roundId {
		roundIdRule = append(roundIdRule, roundIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _HourlySquadGame.contract.WatchLogs(opts, "BetPlaced", roundIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HourlySquadGameBetPlaced)
				if err := _HourlySquadGame.contract.UnpackLog(event, "BetPlaced", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseBetPlaced is a log parse operation binding the contract event 0x4f1eed5e863a822b0f9eb960dfdab2cc5a99beec4b191f2a7a9c7e28e5a15524.
//
// Solidity: event BetPlaced(uint256 indexed roundId, address indexed user, uint8 faction, uint256 amount)
func (_HourlySquadGame *HourlySquadGameFilterer) ParseBetPlaced(log types.Log) (*HourlySquadGameBetPlaced, error) {
	event := new(HourlySquadGameBetPlaced)
	if err := _HourlySquadGame.contract.UnpackLog(event, "BetPlaced", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HourlySquadGameMaxBetAmountUpdatedIterator is returned from FilterMaxBetAmountUpdated and is used to iterate over the raw logs and unpacked data for MaxBetAmountUpdated events raised by the HourlySquadGame contract.
type HourlySquadGameMaxBetAmountUpdatedIterator struct {
	Event *HourlySquadGameMaxBetAmountUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *HourlySquadGameMaxBetAmountUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HourlySquadGameMaxBetAmountUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(HourlySquadGameMaxBetAmountUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *HourlySquadGameMaxBetAmountUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HourlySquadGameMaxBetAmountUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HourlySquadGameMaxBetAmountUpdated represents a MaxBetAmountUpdated event raised by the HourlySquadGame contract.
type HourlySquadGameMaxBetAmountUpdated struct {
	NewAmount *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterMaxBetAmountUpdated is a free log retrieval operation binding the contract event 0x0e858e4fc13b98505f129e4c64b83a3fcdfcfba2bd6b17d1a218390d19d0e769.
//
// Solidity: event MaxBetAmountUpdated(uint256 newAmount)
func (_HourlySquadGame *HourlySquadGameFilterer) FilterMaxBetAmountUpdated(opts *bind.FilterOpts) (*HourlySquadGameMaxBetAmountUpdatedIterator, error) {

	logs, sub, err := _HourlySquadGame.contract.FilterLogs(opts, "MaxBetAmountUpdated")
	if err != nil {
		return nil, err
	}
	return &HourlySquadGameMaxBetAmountUpdatedIterator{contract: _HourlySquadGame.contract, event: "MaxBetAmountUpdated", logs: logs, sub: sub}, nil
}

// WatchMaxBetAmountUpdated is a free log subscription operation binding the contract event 0x0e858e4fc13b98505f129e4c64b83a3fcdfcfba2bd6b17d1a218390d19d0e769.
//
// Solidity: event MaxBetAmountUpdated(uint256 newAmount)
func (_HourlySquadGame *HourlySquadGameFilterer) WatchMaxBetAmountUpdated(opts *bind.WatchOpts, sink chan<- *HourlySquadGameMaxBetAmountUpdated) (event.Subscription, error) {

	logs, sub, err := _HourlySquadGame.contract.WatchLogs(opts, "MaxBetAmountUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HourlySquadGameMaxBetAmountUpdated)
				if err := _HourlySquadGame.contract.UnpackLog(event, "MaxBetAmountUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMaxBetAmountUpdated is a log parse operation binding the contract event 0x0e858e4fc13b98505f129e4c64b83a3fcdfcfba2bd6b17d1a218390d19d0e769.
//
// Solidity: event MaxBetAmountUpdated(uint256 newAmount)
func (_HourlySquadGame *HourlySquadGameFilterer) ParseMaxBetAmountUpdated(log types.Log) (*HourlySquadGameMaxBetAmountUpdated, error) {
	event := new(HourlySquadGameMaxBetAmountUpdated)
	if err := _HourlySquadGame.contract.UnpackLog(event, "MaxBetAmountUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HourlySquadGameOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the HourlySquadGame contract.
type HourlySquadGameOwnershipTransferredIterator struct {
	Event *HourlySquadGameOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *HourlySquadGameOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HourlySquadGameOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(HourlySquadGameOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *HourlySquadGameOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HourlySquadGameOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HourlySquadGameOwnershipTransferred represents a OwnershipTransferred event raised by the HourlySquadGame contract.
type HourlySquadGameOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_HourlySquadGame *HourlySquadGameFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*HourlySquadGameOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _HourlySquadGame.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &HourlySquadGameOwnershipTransferredIterator{contract: _HourlySquadGame.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_HourlySquadGame *HourlySquadGameFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *HourlySquadGameOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _HourlySquadGame.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HourlySquadGameOwnershipTransferred)
				if err := _HourlySquadGame.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_HourlySquadGame *HourlySquadGameFilterer) ParseOwnershipTransferred(log types.Log) (*HourlySquadGameOwnershipTransferred, error) {
	event := new(HourlySquadGameOwnershipTransferred)
	if err := _HourlySquadGame.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HourlySquadGameRewardsClaimedIterator is returned from FilterRewardsClaimed and is used to iterate over the raw logs and unpacked data for RewardsClaimed events raised by the HourlySquadGame contract.
type HourlySquadGameRewardsClaimedIterator struct {
	Event *HourlySquadGameRewardsClaimed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *HourlySquadGameRewardsClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HourlySquadGameRewardsClaimed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(HourlySquadGameRewardsClaimed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *HourlySquadGameRewardsClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HourlySquadGameRewardsClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HourlySquadGameRewardsClaimed represents a RewardsClaimed event raised by the HourlySquadGame contract.
type HourlySquadGameRewardsClaimed struct {
	RoundId *big.Int
	User    common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRewardsClaimed is a free log retrieval operation binding the contract event 0x3300bdb359cfb956935bca32e9db727413eab1ca84341f2e36caea85bb796968.
//
// Solidity: event RewardsClaimed(uint256 indexed roundId, address indexed user, uint256 amount)
func (_HourlySquadGame *HourlySquadGameFilterer) FilterRewardsClaimed(opts *bind.FilterOpts, roundId []*big.Int, user []common.Address) (*HourlySquadGameRewardsClaimedIterator, error) {

	var roundIdRule []interface{}
	for _, roundIdItem := range roundId {
		roundIdRule = append(roundIdRule, roundIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _HourlySquadGame.contract.FilterLogs(opts, "RewardsClaimed", roundIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return &HourlySquadGameRewardsClaimedIterator{contract: _HourlySquadGame.contract, event: "RewardsClaimed", logs: logs, sub: sub}, nil
}

// WatchRewardsClaimed is a free log subscription operation binding the contract event 0x3300bdb359cfb956935bca32e9db727413eab1ca84341f2e36caea85bb796968.
//
// Solidity: event RewardsClaimed(uint256 indexed roundId, address indexed user, uint256 amount)
func (_HourlySquadGame *HourlySquadGameFilterer) WatchRewardsClaimed(opts *bind.WatchOpts, sink chan<- *HourlySquadGameRewardsClaimed, roundId []*big.Int, user []common.Address) (event.Subscription, error) {

	var roundIdRule []interface{}
	for _, roundIdItem := range roundId {
		roundIdRule = append(roundIdRule, roundIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _HourlySquadGame.contract.WatchLogs(opts, "RewardsClaimed", roundIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HourlySquadGameRewardsClaimed)
				if err := _HourlySquadGame.contract.UnpackLog(event, "RewardsClaimed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRewardsClaimed is a log parse operation binding the contract event 0x3300bdb359cfb956935bca32e9db727413eab1ca84341f2e36caea85bb796968.
//
// Solidity: event RewardsClaimed(uint256 indexed roundId, address indexed user, uint256 amount)
func (_HourlySquadGame *HourlySquadGameFilterer) ParseRewardsClaimed(log types.Log) (*HourlySquadGameRewardsClaimed, error) {
	event := new(HourlySquadGameRewardsClaimed)
	if err := _HourlySquadGame.contract.UnpackLog(event, "RewardsClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HourlySquadGameRoundSettledIterator is returned from FilterRoundSettled and is used to iterate over the raw logs and unpacked data for RoundSettled events raised by the HourlySquadGame contract.
type HourlySquadGameRoundSettledIterator struct {
	Event *HourlySquadGameRoundSettled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *HourlySquadGameRoundSettledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HourlySquadGameRoundSettled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(HourlySquadGameRoundSettled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *HourlySquadGameRoundSettledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HourlySquadGameRoundSettledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HourlySquadGameRoundSettled represents a RoundSettled event raised by the HourlySquadGame contract.
type HourlySquadGameRoundSettled struct {
	RoundId            *big.Int
	WinningFaction     uint8
	TotalPool          *big.Int
	WinnerPool         *big.Int
	SettledBlockNumber *big.Int
	SettledBlockHash   [32]byte
	FactionPools       [4]*big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterRoundSettled is a free log retrieval operation binding the contract event 0xc2b8ff06bd5d025468d7b1f59b34f98ac843be81a0cb25edcb0541f4e937add5.
//
// Solidity: event RoundSettled(uint256 indexed roundId, uint8 winningFaction, uint256 totalPool, uint256 winnerPool, uint256 settledBlockNumber, bytes32 settledBlockHash, uint256[4] factionPools)
func (_HourlySquadGame *HourlySquadGameFilterer) FilterRoundSettled(opts *bind.FilterOpts, roundId []*big.Int) (*HourlySquadGameRoundSettledIterator, error) {

	var roundIdRule []interface{}
	for _, roundIdItem := range roundId {
		roundIdRule = append(roundIdRule, roundIdItem)
	}

	logs, sub, err := _HourlySquadGame.contract.FilterLogs(opts, "RoundSettled", roundIdRule)
	if err != nil {
		return nil, err
	}
	return &HourlySquadGameRoundSettledIterator{contract: _HourlySquadGame.contract, event: "RoundSettled", logs: logs, sub: sub}, nil
}

// WatchRoundSettled is a free log subscription operation binding the contract event 0xc2b8ff06bd5d025468d7b1f59b34f98ac843be81a0cb25edcb0541f4e937add5.
//
// Solidity: event RoundSettled(uint256 indexed roundId, uint8 winningFaction, uint256 totalPool, uint256 winnerPool, uint256 settledBlockNumber, bytes32 settledBlockHash, uint256[4] factionPools)
func (_HourlySquadGame *HourlySquadGameFilterer) WatchRoundSettled(opts *bind.WatchOpts, sink chan<- *HourlySquadGameRoundSettled, roundId []*big.Int) (event.Subscription, error) {

	var roundIdRule []interface{}
	for _, roundIdItem := range roundId {
		roundIdRule = append(roundIdRule, roundIdItem)
	}

	logs, sub, err := _HourlySquadGame.contract.WatchLogs(opts, "RoundSettled", roundIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HourlySquadGameRoundSettled)
				if err := _HourlySquadGame.contract.UnpackLog(event, "RoundSettled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoundSettled is a log parse operation binding the contract event 0xc2b8ff06bd5d025468d7b1f59b34f98ac843be81a0cb25edcb0541f4e937add5.
//
// Solidity: event RoundSettled(uint256 indexed roundId, uint8 winningFaction, uint256 totalPool, uint256 winnerPool, uint256 settledBlockNumber, bytes32 settledBlockHash, uint256[4] factionPools)
func (_HourlySquadGame *HourlySquadGameFilterer) ParseRoundSettled(log types.Log) (*HourlySquadGameRoundSettled, error) {
	event := new(HourlySquadGameRoundSettled)
	if err := _HourlySquadGame.contract.UnpackLog(event, "RoundSettled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
