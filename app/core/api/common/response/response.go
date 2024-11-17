package response

import (
	"context"

	"schisandra-album-cloud-microservices/app/core/api/common/i18n"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
)

// SuccessWithData returns a success response with the given data.
func SuccessWithData[T any](data T) *types.Response {
	return &types.Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// Success returns a success response with nil data.
func Success() *types.Response {
	return &types.Response{
		Code:    200,
		Message: "success",
		Data:    nil,
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
func ErrorWithCode(code int64) *types.Response {
	return &types.Response{
		Code:    code,
		Message: "error",
		Data:    nil,
	}
}

// ErrorWithCodeMessage returns an error response with the given code and message.
func ErrorWithCodeMessage(code int64, message string) *types.Response {
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

// ErrorWithI18n returns an error response with the given message.
func ErrorWithI18n(ctx context.Context, msgId string) *types.Response {
	message := i18n.FormatText(ctx, msgId)
	return &types.Response{
		Code:    500,
		Message: message,
		Data:    nil,
	}
}
