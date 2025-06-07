package banktransfer

import "errors"

var WarnError = []error{
	ErrInvalidInput,
	ErrRequiredField,
	ErrOrderNotFound,
	ErrOrderNotCompleted,
	ErrOrderNotPending,
}

var (
	ErrInternal          = errors.New("internal server error")
	ErrInvalidInput      = errors.New("invalid input")
	ErrRequiredField     = errors.New("required field is missing")
	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderNotPending   = errors.New("order is not pending")
	ErrOrderNotCompleted = errors.New("order is not completed")
	ErrInvalidGateway    = errors.New("invalid gateway")
)

func IsWarnError(err error) bool {
	for _, warnError := range WarnError {
		if errors.Is(err, warnError) {
			return true
		}
	}
	return false
}
