package api

// APIResponse is the common envelope for Smart Node HTTP API replies.
// Concrete response types embed it so JSON field names stay `status` and `error`.
type APIResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// APIError returns the API error string, or empty if the call succeeded.
func (r APIResponse) APIError() string { return r.Error }
