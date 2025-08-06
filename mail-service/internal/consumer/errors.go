package consumer

type ValidationError struct {
	Err error
}

func (e *ValidationError) Error() string {
	return "validation error: " + e.Err.Error()
}
