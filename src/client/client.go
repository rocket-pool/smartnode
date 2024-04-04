package client

import (
	"log/slog"

	"github.com/rocket-pool/node-manager-core/api/client"
)

// Binder for the Smart Node daemon API server
type ApiClient struct {
	context  *client.RequesterContext
	Auction  *AuctionRequester
	Faucet   *FaucetRequester
	Minipool *MinipoolRequester
	Network  *NetworkRequester
	Node     *NodeRequester
	ODao     *ODaoRequester
	PDao     *PDaoRequester
	Queue    *QueueRequester
	Security *SecurityRequester
	Service  *ServiceRequester
	Tx       *TxRequester
	Wallet   *WalletRequester
}

// Creates a new API client instance
func NewApiClient(baseRoute string, socketPath string, logger *slog.Logger) *ApiClient {
	context := client.NewRequesterContext(baseRoute, socketPath, logger)

	client := &ApiClient{
		context:  context,
		Auction:  NewAuctionRequester(context),
		Faucet:   NewFaucetRequester(context),
		Minipool: NewMinipoolRequester(context),
		Network:  NewNetworkRequester(context),
		Node:     NewNodeRequester(context),
		ODao:     NewODaoRequester(context),
		PDao:     NewPDaoRequester(context),
		Queue:    NewQueueRequester(context),
		Security: NewSecurityRequester(context),
		Service:  NewServiceRequester(context),
		Tx:       NewTxRequester(context),
		Wallet:   NewWalletRequester(context),
	}
	return client
}
