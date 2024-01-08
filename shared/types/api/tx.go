package api

type TxSignMessageData struct {
	SignedMessage string `json:"signedMessage"`
}

type TxSignTxData struct {
	SignedTx string `json:"signedTx"`
}

type TxBatchSignTxData struct {
	SignedTxs []string `json:"signedTxs"`
}
