package aggregate

import "errors"

// ErrInvalidWindow is returned when a non-positive window duration is supplied.
var ErrInvalidWindow = errors.New("aggregate: window duration must be positive")
