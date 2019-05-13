package nameservice


import (
	"fmt"
	"strings"


	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/tendermint/tendermint/abci/types"
)


// query endpoints supported by the nameservice Querier
const (
	QueryResolve = "resolve"
	QueryWhois = "whois"
	QueryNames = "names"
)


// Module level router for state queries
// noinspection ALL
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error){
		switch path[0] {
		case QueryResolve:
			return queryResolve(ctx, path[1:], req, keeper)
		case QueryWhois:
			return queryWhois(ctx, path[1:], req, keeper)
		case QueryNames:
			return queryNames(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown nameservice query endpoint")
		}
	}
}


// Query resolver
func queryResolve(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	name := path[0]

	value := keeper.resolveName(ctx, name)

	if value == "" {
		return []byte{}, sdk.ErrUnkownRequest("could not resolve name")
	}

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, QueryResResolve{value})
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	//noinspection GoTypesCompatibility
	return bz, nil
}


// Query Result Payload for a resolve query
type QueryResResolve struct {
	Value string `json:"Value"`
}


// fmt.Stringer
func (r QueryResResolve) String() string {
	return r.Value
}


// Query for Whois
func queryWhois(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	name := path[0]

	whois := keeper.GetWhois(ctx, name)

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, whois)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	//noinspection GoTypesCompatibility
	return bz, nil
}

// Query Result Payload for a names query
type QueryResNames []string


// fmt.Stringer
func (n QueryResNames) String() string {
	return strings.Join(n[:], "\n")
}
