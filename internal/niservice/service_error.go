package niservice

// ServiceError is returned when the NI service call failed
type ServiceError struct {
	Message string
	Err     error
}

// Error formats the ServiceError as a printable string
func (e *ServiceError) Error() string {
	return e.Message + ": " + e.Err.Error()
}

// NewServiceError initializes a new error which happened when
// calling the NI service
func NewServiceError(message string, err error) *ServiceError {
	return &ServiceError{
		Message: message,
		Err:     err,
	}
}
