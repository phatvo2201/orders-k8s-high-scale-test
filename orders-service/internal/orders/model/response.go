package model

type Response struct {
	OrderID    string `json:"order_id,omitempty"`
	Message    string `json:"message"`
	Status     string `json:"status"`
	StatusCode int32  `json:"status_code"`
}
