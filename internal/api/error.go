package api

type NotFoundError struct {
}

func (e NotFoundError) Error() string {
	return "could not be found"
}
