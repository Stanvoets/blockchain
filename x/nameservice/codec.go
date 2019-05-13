package nameservice


import (
	"github.com/cosmos/cosmos-sdk/codec"
)


// Registers concrete types on wire codex
func RegisterCodex(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSetName{}, "nameservice/SetName", nil)
	cdc.RegisterConcrete(MsgBuyName{}, "nameservice/BuyName", nil)
}