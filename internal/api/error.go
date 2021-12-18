package api

// NotFoundError is returned when an API resource could not be found.
type NotFoundError struct {
}

func (e NotFoundError) Error() string {
	return "could not be found"
}
