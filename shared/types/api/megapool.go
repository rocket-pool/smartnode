package api

type MegapoolStatusResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Test   bool   `json:"test"`
}
