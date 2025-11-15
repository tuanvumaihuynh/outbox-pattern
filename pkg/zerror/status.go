package zerror

type Status string

const (
	StatusUnknown             Status = "Unknown"
	StatusUnauthorized        Status = "Unauthorized"
	StatusForbidden           Status = "Forbidden"
	StatusNotFound            Status = "Not found"
	StatusUnprocessableEntity Status = "Unprocessable entity"
	StatusConflict            Status = "Conflict"
	StatusTooManyRequests     Status = "Too many requests"
	StatusBadRequest          Status = "Bad request"
	StatusValidationFailed    Status = "Validation failed"
	StatusInternalServerError Status = "Internal server error"
	StatusTimeout             Status = "Timeout"
	StatusNotImplemented      Status = "Not implemented"
	StatusBadGateway          Status = "Bad gateway"
	StatusServiceUnavailable  Status = "Service unavailable"
)

func (s Status) String() string {
	return string(s)
}
