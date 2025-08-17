package types

import "time"

// Response define standard response format
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Ts      string      `json:"ts"`
}

// NewSuccessResponse creates a successful response
func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Code: 0,
		Data: data,
		Ts:   time.Now().Format(time.RFC3339),
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Ts:      time.Now().Format(time.RFC3339),
	}
}

// NewErrorResponseWithRequestID creates an error response with request ID
func NewErrorResponseWithRequestID(code int, message string, requestID string) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Ts:      time.Now().Format(time.RFC3339),
	}
}

// WithMessage adds or updates the message
func (r *Response) WithMessage(message string) *Response {
	r.Message = message
	return r
}
