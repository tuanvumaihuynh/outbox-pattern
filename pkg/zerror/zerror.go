package zerror

import (
	"fmt"
)

// ZError represents the error structure.
type ZError struct {
	parent error
	status Status
	code   string
	msg    string
}

// NewZError initializes a ZError instance.
//
// code example: PRODUCT_NOT_FOUND
func NewZError(parent error, status Status, code, msg string) ZError {
	return ZError{
		parent: parent,
		status: status,
		code:   code,
		msg:    msg,
	}
}

// Error returns the error message for the ZError.
func (e ZError) Error() string {
	if e.parent != nil {
		return fmt.Sprintf("Code=%s, Msg=%s, Parent=(%v)", e.code, e.msg, e.parent)
	}
	return fmt.Sprintf("Code=%s, Msg=%s", e.code, e.msg)
}

// WrapParent attaches an underlying error to an existing predefined ZError.
func (e ZError) WrapParent(parent error) ZError {
	if parent == nil {
		return e
	}
	e.parent = parent
	return e
}

// Unwrap returns the underlying error for the ZError.
func (e *ZError) Unwrap() error {
	return e.parent
}

// Status returns the status of the ZError.
func (e ZError) Status() Status {
	return e.status
}

// Code returns the code of the ZError.
func (e ZError) Code() string {
	return e.code
}

// Msg returns the message of the ZError.
func (e ZError) Msg() string {
	return e.msg
}

// Parent returns the underlying error for the ZError.
func (e ZError) Parent() error {
	return e.parent
}

func NewUnauthorized(code, msg string) ZError {
	return NewZError(nil, StatusUnauthorized, code, msg)
}

func NewForbidden(code, msg string) ZError {
	return NewZError(nil, StatusForbidden, code, msg)
}

func NewNotFound(code, msg string) ZError {
	return NewZError(nil, StatusNotFound, code, msg)
}

func NewUnprocessableEntity(code, msg string) ZError {
	return NewZError(nil, StatusUnprocessableEntity, code, msg)
}

func NewConflict(code, msg string) ZError {
	return NewZError(nil, StatusConflict, code, msg)
}

func NewTooManyRequests(code, msg string) ZError {
	return NewZError(nil, StatusTooManyRequests, code, msg)
}

func NewBadRequest(code, msg string) ZError {
	return NewZError(nil, StatusBadRequest, code, msg)
}

func NewValidationFailed(code, msg string) ZError {
	return NewZError(nil, StatusValidationFailed, code, msg)
}

func NewInternalServerError(code, msg string) ZError {
	return NewZError(nil, StatusInternalServerError, code, msg)
}

func NewTimeout(code, msg string) ZError {
	return NewZError(nil, StatusTimeout, code, msg)
}

func NewNotImplemented(code, msg string) ZError {
	return NewZError(nil, StatusNotImplemented, code, msg)
}

func NewBadGateway(code, msg string) ZError {
	return NewZError(nil, StatusBadGateway, code, msg)
}

func NewServiceUnavailable(code, msg string) ZError {
	return NewZError(nil, StatusServiceUnavailable, code, msg)
}
