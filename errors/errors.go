package errors

import (
	"github.com/morikuni/failure"
)

// Error code
const (
	Internal         failure.StringCode = "Internal"
	Unauthenticated  failure.StringCode = "Unauthenticated"
	InvalidArguments failure.StringCode = "InvalidArguments"
	NotFoundEntity   failure.StringCode = "NotFoundEntity"
	NotFailureError  failure.StringCode = "NotFailureError"
)

func GetMessage(err error) string {
	if err == nil {
		return ""
	}
	if message, ok := failure.MessageOf(err); ok {
		return message
	}
	if code, ok := failure.CodeOf(err); ok {
		return code.ErrorCode()
	}
	return err.Error()
}

func IsFailureError(err error) bool {
	_, hasCode := failure.CodeOf(err)
	return hasCode
}
