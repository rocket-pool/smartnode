package rocketpool

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

const (
	baseUrl         string = "http://rocketpool/%s"
	jsonContentType string = "application/json"
)

type IRequester interface {
	GetName() string
	GetRoute() string
	GetClient() *http.Client
}

// Binder for the Rocket Pool daemon API server
type ApiRequester struct {
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

	socketPath string
	client     *http.Client
}

// Creates a new API requester instance
func NewApiRequester(socketPath string) *ApiRequester {
	apiRequester := &ApiRequester{
		socketPath: socketPath,
	}
	apiRequester.client = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	apiRequester.Auction = NewAuctionRequester(apiRequester.client)
	apiRequester.Faucet = NewFaucetRequester(apiRequester.client)
	apiRequester.Minipool = NewMinipoolRequester(apiRequester.client)
	apiRequester.Network = NewNetworkRequester(apiRequester.client)
	apiRequester.Node = NewNodeRequester(apiRequester.client)
	apiRequester.ODao = NewODaoRequester(apiRequester.client)
	apiRequester.PDao = NewPDaoRequester(apiRequester.client)
	apiRequester.Queue = NewQueueRequester(apiRequester.client)
	apiRequester.Security = NewSecurityRequester(apiRequester.client)
	apiRequester.Service = NewServiceRequester(apiRequester.client)
	apiRequester.Tx = NewTxRequester(apiRequester.client)
	apiRequester.Wallet = NewWalletRequester(apiRequester.client)
	return apiRequester
}

// Submit a GET request to the API server
func sendGetRequest[DataType any](r IRequester, method string, requestName string, args map[string]string) (*api.ApiResponse[DataType], error) {
	if args == nil {
		args = map[string]string{}
	}
	response, err := rawGetRequest[DataType](r.GetClient(), fmt.Sprintf("%s/%s", r.GetRoute(), method), args)
	if err != nil {
		return nil, fmt.Errorf("error during %s %s request: %w", r.GetName(), requestName, err)
	}
	return response, nil
}

// Submit a GET request to the API server
func rawGetRequest[DataType any](client *http.Client, path string, params map[string]string) (*api.ApiResponse[DataType], error) {
	// Create the request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(baseUrl, path), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Encode the params into a query string
	values := url.Values{}
	for name, value := range params {
		values.Add(name, value)
	}
	req.URL.RawQuery = values.Encode()

	// Run the request
	resp, err := client.Do(req)
	return handleResponse[DataType](resp, path, err)
}

// Submit a POST request to the API server
func sendPostRequest[DataType any](r IRequester, method string, requestName string, body any) (*api.ApiResponse[DataType], error) {
	// Serialize the body
	bytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error serializing request body for %s %s: %w", r.GetName(), requestName, err)
	}

	response, err := rawPostRequest[DataType](r.GetClient(), fmt.Sprintf("%s/%s", r.GetRoute(), method), string(bytes))
	if err != nil {
		return nil, fmt.Errorf("error during %s %s request: %w", r.GetName(), requestName, err)
	}
	return response, nil
}

// Submit a POST request to the API server
func rawPostRequest[DataType any](client *http.Client, path string, body string) (*api.ApiResponse[DataType], error) {
	resp, err := client.Post(fmt.Sprintf(baseUrl, path), jsonContentType, strings.NewReader(body))
	return handleResponse[DataType](resp, path, err)
}

// Processes a response to a request
func handleResponse[DataType any](resp *http.Response, path string, err error) (*api.ApiResponse[DataType], error) {
	if err != nil {
		return nil, fmt.Errorf("error requesting %s: %w", path, err)
	}

	// Read the body
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)

	// Check if the request failed
	if resp.StatusCode != http.StatusOK {
		if err != nil {
			return nil, fmt.Errorf("server responded to %s with code %s but reading the response body failed: %w", path, resp.Status, err)
		}
		msg := string(bytes)
		return nil, fmt.Errorf("server responded to %s with code %s: [%s]", path, resp.Status, msg)
	}
	if err != nil {
		return nil, fmt.Errorf("error reading the response body for %s: %w", path, err)
	}

	// Deserialize the response into the provided type
	var parsedResponse api.ApiResponse[DataType]
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("error deserializing response to %s: %w; original body: [%s]", path, err, string(bytes))
	}

	return &parsedResponse, nil
}

// Types that can be batched into a comma-delmited string
type BatchInputType interface {
	uint64 | common.Address
}

// Converts an array of inputs into a comma-delimited string
func makeBatchArg[DataType BatchInputType](input []DataType) string {
	results := make([]string, len(input))

	// Figure out how to stringify the input
	switch typedInput := any(&input).(type) {
	case *[]uint64:
		for i, index := range *typedInput {
			results[i] = strconv.FormatUint(index, 10)
		}
	case *[]common.Address:
		for i, address := range *typedInput {
			results[i] = address.Hex()
		}
	}
	return strings.Join(results, ",")
}
