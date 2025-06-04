package model

type BaseResponse struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message"`
}

type EventCreateRequest struct {
	Name string `json:"name"`
}
