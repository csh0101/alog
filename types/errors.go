package types

import "errors"

var (
	// ErrUnsupported will be thrown when functions are unimplemented.
	ErrUnsupported = errors.New("unsupported")
)
