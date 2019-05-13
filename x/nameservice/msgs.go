package nameservice

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*
---------------------------------- SET NAME ----------------------------------
*/

// MsgSetName defines a SetName message
type MsgSetName struct {
	Name string
	Value string
	Owner sdk.AccAddress
}


// NewMsgSetName is a constructor function for MsgSetName
func NewMsgSetName (name string, value string, owner sdk.AccAddress) MsgSetName {
	return MsgSetName{
		Name: name,
		Value: value,
		Owner: owner,
	}
}


// Route returns name of the module this type belongs to.
func (msg MsgSetName) Route() string { return "nameservice" }


// Returns the action
func (msg MsgSetName) Type() string { return "set_name" }


// Runs stateless checks on the message
func (msg MsgSetName) ValidateBasic() sdk.Error {
	if msg.Owner.Empty() {
		return sdk.ErrInvalidAddress(msg.Owner.String())
	}
	if len(msg.Name) == 0 || len(msg.value) == 0 {
		return sdk.ErrUnknownRequest("Name and/or Value can't be empty")
	}
	return nil
}


// Encodes the message for signing
func (msg MsgSetName) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}


// Defines whose signature is required
func (msg MsgSetName) GetSingers() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}
/*
------------------------------------------------------------------------------
*/


/*
---------------------------------- BUY NAME ----------------------------------
*/

// Buy name message
type MsgBuyName struct {
	Name string
	Bid sdk.Coins
	Buyer sdk.AccAddress
}


// Constructor function for MsgBuyName
func NewMsgBuyName(name string, bid sdk.Coins, buyer sdk.AccAddress) MsgBuyName {
	return MsgBuyName{
		Name: name,
		Bid: bid,
		Buyer: buyer,
	}
}


// Return the name of the module
func (msg MsgBuyName) Route() string {
	return "nameservice"
}


// Return the name of the action
func (msg MsgBuyName) Type() string {
	return "buy_name"
}


// Runs stateless checks on the message
func (msg MsgBuyName) ValidateBasic() sdk.Error {
	if msg.Buyer.Empty() {
		return sdk.ErrInvalidAddress(msg.Buyer.String())
	}
	if len(msg.Name) == 0 {
		return sdk.ErrUnkownRequest("Name can't be empty")
	}
	if !msg.Bid.IsAllPositive() {
		return sdk.ErrInsufficientCoins("Bids must be positive")
	}
	return nil
}


// Encodes the message for signing
func (msg MsgBuyName) GetSignBytes() []byte {
	b,err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}


// Defines whose signature is required
func (msg MsgBuyName) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress(msg.Buyer)
}
/*
------------------------------------------------------------------------------
*/