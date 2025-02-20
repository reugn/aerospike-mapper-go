package mapper

import "errors"

var (
	ErrInvalidSource     = errors.New("source does not contain aerospike record")
	ErrInvalidSourceType = errors.New("source must be a struct or a pointer to a struct")
)
