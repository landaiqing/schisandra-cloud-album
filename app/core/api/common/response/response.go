package response

import "schisandra-album-cloud-microservices/app/core/api/internal/types"

// Success returns a success response with the given data.
func Success[T any](data T) *types.Response {
	return &types.Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// SuccessWithMessage returns a success response with the given message.
func SuccessWithMessage(message string) *types.Response {
	return &types.Response{
		Code:    200,
		Message: message,
		Data:    nil,
	}
}

// Error returns an error response with the given message.
func Error() *types.Response {
	return &types.Response{
		Code:    500,
		Message: "error",
		Data:    nil,
	}
}

// ErrorWithCode returns an error response with the given code and message.
func ErrorWithCode(code int64, message string) *types.Response {
	return &types.Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// ErrorWithMessage returns an error response with the given message.
func ErrorWithMessage(message string) *types.Response {
	return &types.Response{
		Code:    500,
		Message: message,
		Data:    nil,
	}
}
