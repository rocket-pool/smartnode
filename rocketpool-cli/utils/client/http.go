package client

import (
	"fmt"
	"io"
	"net/http"

	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

const (
	baseUrl string = "http://node/%s"
)

func SendGetRequest[DataType any](client *Client, path string) (api.ApiResponse[DataType], error) {
	var parsedResponse api.ApiResponse[DataType]

	// Run the request
	resp, err := client.apiClient.Get(fmt.Sprintf(baseUrl, path))
	if err != nil {
		return parsedResponse, fmt.Errorf("error requesting %s: %w", path, err)
	}
	defer resp.Body.Close()

	// Read the body
	bytes, err := io.ReadAll(resp.Body)

	// Check if the request failed
	if resp.StatusCode != http.StatusOK {
		if err != nil {
			return parsedResponse, fmt.Errorf("server responded to %s with code %s but reading the response body failed: %w", path, resp.Status, err)
		}
		msg := string(bytes)
		return parsedResponse, fmt.Errorf("server responded to %s with code %s: [%s]", path, resp.Status, msg)
	}
	if err != nil {
		return parsedResponse, fmt.Errorf("error reading the response body for %s: %w", path, err)
	}

	// Deserialize the response into the provided type
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		return parsedResponse, fmt.Errorf("error deserializing response to %s: %w; original body: [%s]", path, err, string(bytes))
	}

	return parsedResponse, nil
}
