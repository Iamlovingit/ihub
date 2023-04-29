package api

// Reply .
type Reply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}
