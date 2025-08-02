package consumer

type ValidationError struct {
	Err error
}

func (e *ValidationError) Error() string {
	return "validation error: " + e.Err.Error()
}

type TemporaryError struct {
	Err error
}

func (e *TemporaryError) Error() string {
	return "temporary error: " + e.Err.Error()
}
