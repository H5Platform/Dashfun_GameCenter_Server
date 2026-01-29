package api

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/nolandev/squadgame"
	"dashfun_gamecenter/web"
	"net/http"
	"strconv"

	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
)

func apiSquadGameClaimRewards(c *gin.Context, user *data.DashFunUser) {
	// Try to find a valid EVM address
	// Get address from request
	address, existed := c.GetPostForm("address")
	if !existed || address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("address is required"))
		return
	}

	txHash, err := squadgame.GetSquadGameService().ClaimUserRewards(address)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(txHash))
}

func apiSquadGamePlaceBet(c *gin.Context, user *data.DashFunUser) {
	// Parsers
	// Faction
	factionStr := c.PostForm("faction")
	// Amount
	amountStr := c.PostForm("amount")
	// Deadline
	deadlineStr := c.PostForm("deadline")
	// Signature parts
	vStr := c.PostForm("v")
	rStr := c.PostForm("r")
	sStr := c.PostForm("s")

	if factionStr == "" || amountStr == "" || deadlineStr == "" || vStr == "" || rStr == "" || sStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("missing parameters"))
		return
	}

	// Conversion
	// Faction: uint8
	factionInt, err := strconv.Atoi(factionStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("invalid faction"))
		return
	}
	faction := uint8(factionInt)

	// Amount: *big.Int (from string)
	amount := new(big.Int)
	amount, ok := amount.SetString(amountStr, 10)
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("invalid amount"))
		return
	}

	// Deadline: *big.Int (from string)
	deadline := new(big.Int)
	deadline, ok = deadline.SetString(deadlineStr, 10)
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("invalid deadline"))
		return
	}

	// V: uint8
	vInt, err := strconv.Atoi(vStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("invalid v"))
		return
	}
	v := uint8(vInt)

	// R: [32]byte (from hex)
	rBytes, err := hexutil.Decode(rStr)
	if err != nil || len(rBytes) != 32 {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("invalid r"))
		return
	}
	var r [32]byte
	copy(r[:], rBytes)

	// S: [32]byte (from hex)
	sBytes, err := hexutil.Decode(sStr)
	if err != nil || len(sBytes) != 32 {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("invalid s"))
		return
	}
	var s [32]byte
	copy(s[:], sBytes)

	// Address
	address := c.PostForm("address")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("missing address"))
		return
	}

	req := squadgame.UserBetRequest{
		Faction:     faction,
		Amount:      amount,
		Deadline:    deadline,
		V:           v,
		R:           r,
		S:           s,
		FromAddress: address,
	}

	txHash, err := squadgame.GetSquadGameService().RelayBet(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(txHash))
}

func apiSquadGameInfo(c *gin.Context, user *data.DashFunUser) {
	address, existed := c.GetPostForm("address")
	if !existed || address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("address is required"))
		return
	}

	bet, err := squadgame.GetSquadGameService().GetUserBetInfo(address)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	// If bet is nil, it means user has never played. Return null data or empty object?
	// RSuccess handles interface{}, typically returns "data": null
	c.JSON(http.StatusOK, RSuccess(bet))
}

func apiGetSquadGameRound(c *gin.Context, user *data.DashFunUser) {
	round, err := squadgame.GetSquadGameService().GetLatestRound()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(round))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleSquadGame, web.POST, "claim-rewards", userHandlerAuthWrapper(apiSquadGameClaimRewards))
	web.GetService().RegisterApi(web.ApiModuleSquadGame, web.POST, "place-bet", userHandlerAuthWrapper(apiSquadGamePlaceBet))
	web.GetService().RegisterApi(web.ApiModuleSquadGame, web.POST, "info", userHandlerAuthWrapper(apiSquadGameInfo))
	web.GetService().RegisterApi(web.ApiModuleSquadGame, web.GET, "round", userHandlerAuthWrapper(apiGetSquadGameRound))
}
