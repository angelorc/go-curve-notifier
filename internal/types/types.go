package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/abci/types"
)

type Telegram struct {
	BotToken string `toml:"bot_token"`
	ChatID   int64  `toml:"chat_id"`
	ThreadID int64  `toml:"thread_id"`
}

type Config struct {
	GRPCServerAddress string   `toml:"grpc_server_address"`
	RPCServerAddress  string   `toml:"rpc_server_address"`
	WebsocketPath     string   `toml:"websocket_path"`
	Telegram          Telegram `toml:"telegram"`
}

type MarketplaceEventActivity struct {
	Action             string   `json:"action"`
	ContractAddress    string   `json:"_contract_address"`
	NftAddress         string   `json:"nft_address"`
	Royalties          sdk.Coin `json:"royalties"`
	RoyaltiesRecipient string   `json:"royalties_recipient"`
	ProtocolFee        sdk.Coin `json:"protocol_fee"`
	Refund             sdk.Coin `json:"refund"`
	Referral           string   `json:"referral"`
	ReferralAmount     sdk.Coin `json:"referral_amount"`
	TokenIds           string   `json:"token_ids"`
	Price              sdk.Coin `json:"price"`
	Burner             string   `json:"burner"`
	Recipient          string   `json:"recipient"`
}

func mustParseCoin(coinStr string) sdk.Coin {
	coin, err := sdk.ParseCoinNormalized(coinStr)
	if err != nil {
		panic(err)
	}
	return coin
}

func NewMarketplaceEventActivityFromAttr(attrs []types.EventAttribute, nftAddress string) *MarketplaceEventActivity {
	var event MarketplaceEventActivity
	var denom string
	defaultDenom := "ubtsg"
	zeroCoin := sdk.NewCoin(defaultDenom, sdk.ZeroInt())

	event.NftAddress = nftAddress
	event.Royalties = zeroCoin
	event.ProtocolFee = zeroCoin
	event.Refund = zeroCoin
	event.ReferralAmount = zeroCoin
	event.Price = zeroCoin

	for _, attr := range attrs {
		switch string(attr.Key) {
		case "action":
			event.Action = string(attr.Value)
		case "_contract_address":
			event.ContractAddress = string(attr.Value)
		case "royalties":
			event.Royalties = mustParseCoin(fmt.Sprintf("%s%s", attr.Value, defaultDenom))
		case "royalties_recipient":
			event.RoyaltiesRecipient = string(attr.Value)
		case "protocol_fee":
			event.ProtocolFee = mustParseCoin(fmt.Sprintf("%s%s", attr.Value, defaultDenom))
		case "refund":
			event.Refund = mustParseCoin(fmt.Sprintf("%s%s", attr.Value, defaultDenom))
		case "referral":
			event.Referral = string(attr.Value)
		case "referral_amount":
			event.ReferralAmount = mustParseCoin(fmt.Sprintf("%s%s", attr.Value, defaultDenom))
		case "token_ids":
			event.TokenIds = string(attr.Value)
		case "price":
			event.Price = mustParseCoin(fmt.Sprintf("%s%s", attr.Value, defaultDenom))
		case "denom":
			denom = string(attr.Value)
		case "burner":
			event.Burner = string(attr.Value)
		case "recipient":
			event.Recipient = string(attr.Value)
		}
	}

	if denom != "" {
		event.Royalties.Denom = denom
		event.ProtocolFee.Denom = denom
		event.Refund.Denom = denom
		event.ReferralAmount.Denom = denom
		event.Price.Denom = denom
	} else {
		panic("denom not found")
	}

	return &event
}

func (m *MarketplaceEventActivity) String() string {
	return "MarketplaceEventActivity{" +
		"Action: " + m.Action + ", " +
		"ContractAddress: " + m.ContractAddress + ", " +
		"NftAddress: " + m.NftAddress + ", " +
		"Royalties: " + m.Royalties.String() + ", " +
		"RoyaltiesRecipient: " + m.RoyaltiesRecipient + ", " +
		"ProtocolFee: " + m.ProtocolFee.String() + ", " +
		"Refund: " + m.Refund.String() + ", " +
		"Referral: " + m.Referral + ", " +
		"ReferralAmount: " + m.ReferralAmount.String() + ", " +
		"TokenIds: " + m.TokenIds + ", " +
		"Price: " + m.Price.String() + ", " +
		"Burner: " + m.Burner + ", " +
		"Recipient: " + m.Recipient + "}"
}
