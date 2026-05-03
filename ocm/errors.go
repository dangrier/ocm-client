package ocm

import "errors"

var (
	ErrLocationNotFound = errors.New("location not found")
	ErrInvalidDateRange = errors.New("date_from must be before date_to")
)
