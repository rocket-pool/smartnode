package client

import (
	"log/slog"
	"net/http/httptrace"
	"net/url"

	"github.com/rocket-pool/node-manager-core/api/client"
)

// Binder for the Smart Node daemon API server
type ApiClient struct {
	context  client.IRequesterContext
	Auction  *AuctionRequester
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
func NewApiClient(apiUrl *url.URL, logger *slog.Logger, tracer *httptrace.ClientTrace) *ApiClient {
	context := client.NewNetworkRequesterContext(apiUrl, logger, tracer, nil)

	client := &ApiClient{
		context:  context,
		Auction:  NewAuctionRequester(context),
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
