package api

// Reply .
type Reply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type ValiReply struct {
	Code int               `json:"code"`
	Data map[string]string `json:"data"`
}
