package srverrors

import "errors"

var (
	ErrorNotFound      = errors.New("resource not found")
	ErrorAlreadyExists = errors.New("resource already exists")
	ErrorTokenNotFound = errors.New("token not found")
	ErrCacheMiss       = errors.New("key not found")
)
