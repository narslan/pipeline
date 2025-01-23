package dataflow

import (
	"errors"
	"fmt"
)

// Application error codes.
const (
	EINTERNAL = "internal"
	EINVALID  = "invalid"
	ENOTFOUND = "not_found"
)

// Error represents an application error.
// Any error unrelated to application functionality, like a disk error,
// should be an EINTERNAL error. For the user, only a generic message
// such as 'Internal error' will be displayed.
// The specific details of these low-level
// errors should be logged and forwarded to the administrator"

type Error struct {
	// Application error code.
	Code string

	// Human-readable error expression.
	Message string
}

// ErrorCode unwraps an application error and returns its code.
// Non-application specific errors always return EINTERNAL.
func ErrorCode(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Code
	}
	return EINTERNAL
}

// ErrorMessage returns its message.
// Non-application specific errors always return "Internal error".
func ErrorMessage(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Message
	}
	return "Internal error."
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("dataflow error: code=%s message=%s", e.Code, e.Message)
}

// Errorf returns an Error with a given code and formatted message.
func Errorf(code string, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}
