package schemas

type Request struct {
	Id      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Jsonrpc string      `json:"jsonrpc"`
}

type Response struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  *struct {
		Value     string `json:"value"`
		SessionId string `json:"sessionId"`
	} `json:"result,omitempty"`
	Error *Error `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}
