package apierr

import (
	"errors"
	"net/http"

	govalidator "github.com/go-playground/validator/v10"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/http/gen"
	"github.com/tuanvumaihuynh/outbox-pattern/pkg/validator"
	"github.com/tuanvumaihuynh/outbox-pattern/pkg/zerror"
)

// ErrorResponse is the error response for the API.
type ErrorResponse struct {
	gen.ErrorResponse

	// StatusCode is the status code for the error response.
	StatusCode int `json:"-"`
}

func New(err error) ErrorResponse {
	return errorToErrorResponse(err)
}

var InternalServerErr = ErrorResponse{
	ErrorResponse: gen.ErrorResponse{
		Code:    "internalServerError",
		Message: "an unknown error occurred",
	},
	StatusCode: http.StatusInternalServerError,
}

func errorToErrorResponse(err error) ErrorResponse {
	var zErr zerror.ZError
	if errors.As(err, &zErr) {
		return ErrorResponse{
			ErrorResponse: gen.ErrorResponse{
				Code:    zErr.Code(),
				Message: zErr.Msg(),
			},
			StatusCode: ZErrorStatusToHTTPStatus(zErr.Status()),
		}
	}

	var validationErrs govalidator.ValidationErrors
	if errors.As(err, &validationErrs) {
		details := make([]gen.FieldError, len(validationErrs))
		for i, fe := range validationErrs {
			details[i] = gen.FieldError{
				Field:   fe.Field(),
				Message: validator.ValidationErrorMessage(fe),
			}
		}

		return ErrorResponse{
			ErrorResponse: gen.ErrorResponse{
				Code:    "validationError",
				Message: "validation error",
				Details: &details,
			},
			StatusCode: http.StatusBadRequest,
		}
	}

	if isOpenAPICodegenErr(err) {
		return ErrorResponse{
			ErrorResponse: gen.ErrorResponse{
				Code:    "validationError",
				Message: err.Error(),
			},
			StatusCode: http.StatusBadRequest,
		}
	}

	return InternalServerErr
}

func ZErrorStatusToHTTPStatus(status zerror.Status) int {
	switch status {
	case zerror.StatusUnauthorized:
		return http.StatusUnauthorized
	case zerror.StatusForbidden:
		return http.StatusForbidden
	case zerror.StatusNotFound:
		return http.StatusNotFound
	case zerror.StatusUnprocessableEntity:
		return http.StatusUnprocessableEntity
	case zerror.StatusConflict:
		return http.StatusConflict
	case zerror.StatusTooManyRequests:
		return http.StatusTooManyRequests
	case zerror.StatusBadRequest:
		return http.StatusBadRequest
	case zerror.StatusValidationFailed:
		return http.StatusBadRequest
	case zerror.StatusUnknown, zerror.StatusInternalServerError:
		return http.StatusInternalServerError
	case zerror.StatusTimeout:
		return http.StatusGatewayTimeout
	case zerror.StatusNotImplemented:
		return http.StatusNotImplemented
	case zerror.StatusBadGateway:
		return http.StatusBadGateway
	case zerror.StatusServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func isOpenAPICodegenErr(err error) bool {
	var (
		e1 *gen.UnescapedCookieParamError
		e2 *gen.UnmarshalingParamError
		e3 *gen.RequiredParamError
		e4 *gen.RequiredHeaderError
		e5 *gen.InvalidParamFormatError
		e6 *gen.TooManyValuesForParamError
	)

	return errors.As(err, &e1) ||
		errors.As(err, &e2) ||
		errors.As(err, &e3) ||
		errors.As(err, &e4) ||
		errors.As(err, &e5) ||
		errors.As(err, &e6)
}
