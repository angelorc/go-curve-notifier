package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/angelorc/go-curve-notifier/internal/config"
	"github.com/angelorc/go-curve-notifier/internal/logger"
	"github.com/angelorc/go-curve-notifier/internal/query"
	"github.com/angelorc/go-curve-notifier/internal/telegram"
	"github.com/angelorc/go-curve-notifier/internal/types"
	tmabcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"strconv"
	"strings"
	"time"
)

func formatCoin(coin sdk.Coin) string {
	amt := coin.Amount.ToLegacyDec().QuoInt64(1_000_000).MustFloat64()
	var denom string
	if coin.Denom == "ubtsg" {
		denom = "BTSG"
	} else {
		denom = coin.Denom
	}

	formatted := strconv.FormatFloat(amt, 'f', 6, 64)

	return fmt.Sprintf("%s %s", formatted, denom)
}

func formatAction(action string) string {
	switch action {
	case "mint_bs721_curve_nft":
		return "minted"
	case "burn_bs721_curve_nft":
		return "burned"
	default:
		return action
	}
}

func handleActivity(l *zap.Logger, cfg *types.Config, qc *query.QueryClient, txHash string, activity *types.MarketplaceEventActivity) {
	nftInfo, err := qc.NftContractInfo(activity.NftAddress)
	if err != nil {
		log.Fatalf("Failed to get NFT contract info: %v", err)
	}

	var account string
	var icons string
	if activity.Action == "mint_bs721_curve_nft" {
		account = activity.Recipient
		icons = "🤑🤑🤑🤑🤑🤑🤑🤑🤑🤑"
	} else {
		account = activity.Burner
		icons = "🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥"
	}

	totalTokens := len(strings.Split(activity.TokenIds, ","))
	text := fmt.Sprintf(`%s

👤 <b><i><a href="https://bitsong.studio/u/%s">%s</a> %s %d <b><a href="https://bitsong.studio/nfts/%s">%s</a></b></i></b>

<b>Royalties</b>: %s
<b>Protocol Fee</b>: %s
<b>Referral</b>: %s
<b>Total Price</b>: %s

<b>Token Ids</b>: %s

<a href="https://mintscan.io/bitsong/tx/%s">%s</a>`,
		icons,
		account,
		account,
		formatAction(activity.Action),
		totalTokens,
		activity.NftAddress,
		nftInfo.Name,
		formatCoin(activity.Royalties),
		formatCoin(activity.ProtocolFee),
		formatCoin(activity.ReferralAmount),
		formatCoin(activity.Price),
		activity.TokenIds,
		txHash,
		txHash,
	)

	err = telegram.SendMessage(cfg, text)
	if err != nil {
		l.Error("Failed to send message", zap.Error(err))
	}

	l.Info("Sent message",
		zap.String("sender", account),
		zap.String("action", activity.Action),
		zap.String("total_price", formatCoin(activity.Price)),
	)
}

func handleEvent(l *zap.Logger, cfg *types.Config, qc *query.QueryClient, event ctypes.ResultEvent) {
	data := event.Data.(tmtypes.EventDataTx)
	events := data.Result.Events
	txHash := fmt.Sprintf("%X", tmhash.Sum(data.TxResult.Tx))

	var marketplaceAttrs []tmabcitypes.EventAttribute
	var nftAddress string

	for _, evt := range events {
		if evt.Type == "wasm" {
			for _, attr := range evt.Attributes {
				if string(attr.Key) == "action" && string(attr.Value) == "mint_bs721_curve_nft" ||
					string(attr.Key) == "action" && string(attr.Value) == "burn_bs721_curve_nft" {
					marketplaceAttrs = evt.Attributes
				}

				if string(attr.Key) == "action" && string(attr.Value) == "mint" ||
					string(attr.Key) == "action" && string(attr.Value) == "burn" {
					for _, attr := range evt.Attributes {
						if string(attr.Key) == "_contract_address" {
							nftAddress = string(attr.Value)
						}
					}
				}
			}
		}
	}

	if len(marketplaceAttrs) > 0 && nftAddress != "" {
		handleActivity(l, cfg, qc, txHash, types.NewMarketplaceEventActivityFromAttr(marketplaceAttrs, nftAddress))
	}
}

func main() {
	configPath := flag.String("c", "config.toml", "path to config file")
	flag.Parse()

	ctx := context.Background()

	l, err := logger.Setup()
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		l.Fatal("Failed to load config", zap.Error(err))
	}

	conn, err := grpc.NewClient(cfg.GRPCServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Fatal("Failed to connect to GRPC server", zap.Error(err))
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			l.Fatal("Error closing GRPC connection",
				zap.Error(err),
			)
		}
	}(conn)

	qc, err := query.NewQueryClient(conn)
	if err != nil {
		l.Fatal("Error initialising query client",
			zap.Error(err),
		)
	}

	startWebsocketClient := func() (*rpchttp.HTTP, error) {
		wsClient, err := rpchttp.New(cfg.RPCServerAddress, cfg.WebsocketPath)
		if err != nil {
			return nil, err
		}

		err = wsClient.Start()
		if err != nil {
			return nil, err
		}
		return wsClient, nil
	}

	subscribeToEvents := func(wsClient *rpchttp.HTTP) (<-chan ctypes.ResultEvent, <-chan ctypes.ResultEvent, error, error) {
		queryMint := fmt.Sprintf("wasm.action = 'mint_bs721_curve_nft'")
		//queryMint := fmt.Sprintf("tm.event='NewBlock'")
		queryBurn := fmt.Sprintf("wasm.action = 'burn_bs721_curve_nft'")

		eventMintCh, errMintCh := wsClient.Subscribe(ctx, "bot", queryMint)
		if errMintCh != nil {
			return nil, nil, errMintCh, nil
		}
		l.Info("Subscribed to mint events", zap.String("query", queryMint))

		eventBurnCh, errBurnCh := wsClient.Subscribe(ctx, "bot", queryBurn)
		if errBurnCh != nil {
			return nil, nil, nil, errBurnCh
		}
		l.Info("Subscribed to burn events", zap.String("query", queryBurn))
		return eventMintCh, eventBurnCh, errMintCh, errBurnCh
	}

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		wsClient, err := startWebsocketClient()
		if err != nil {
			l.Fatal("Error starting websocket client", zap.Error(err))
		}

		eventMintCh, eventBurnCh, errMintCh, errBurnCh := subscribeToEvents(wsClient)
		if errMintCh != nil || errBurnCh != nil {
			l.Fatal("Error subscribing websocket client", zap.Errors("errors", []error{errMintCh, errBurnCh}))
		}

		for {
			select {
			case event := <-eventMintCh:
				//fmt.Println("Mint event", event)
				handleEvent(l, cfg, qc, event)
			case event := <-eventBurnCh:
				//fmt.Println("Burn event")
				handleEvent(l, cfg, qc, event)
			case <-ticker.C:
				//fmt.Println("Reconnecting websocket client")
				err := wsClient.Stop()
				if err != nil {
					l.Fatal("Error stopping websocket client", zap.Error(err))
				}
				wsClient, err = startWebsocketClient()
				if err != nil {
					l.Fatal("Error restarting websocket client", zap.Error(err))
				}
				eventMintCh, eventBurnCh, errMintCh, errBurnCh = subscribeToEvents(wsClient)
				if errMintCh != nil || errBurnCh != nil {
					l.Fatal("Error re-subscribing websocket client", zap.Errors("errors", []error{errMintCh, errBurnCh}))
				}
			}
		}
	}()

	select {}
}
