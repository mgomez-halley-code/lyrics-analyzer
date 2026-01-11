package handler

// NotFoundError represents an error when requested resource is not found
type NotFoundError struct {
	Err error
}

func (e *NotFoundError) Error() string {
	return e.Err.Error()
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

// RateLimitError represents an error when rate limit is exceeded
type RateLimitError struct {
	Err error
}

func (e *RateLimitError) Error() string {
	return e.Err.Error()
}

func (e *RateLimitError) Unwrap() error {
	return e.Err
}

// TimeoutError represents an error when request times out
type TimeoutError struct {
	Err error
}

func (e *TimeoutError) Error() string {
	return e.Err.Error()
}

func (e *TimeoutError) Unwrap() error {
	return e.Err
}
