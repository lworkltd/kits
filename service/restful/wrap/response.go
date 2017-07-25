package wrap

type Response struct {
	Result  bool        `json:"result"`
	Mcode   string      `json:"mcode"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
