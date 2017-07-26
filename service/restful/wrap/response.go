package wrap

type Response struct {
	Result  bool        `json:"result"`
	Mcode   string      `json:"mcode,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
