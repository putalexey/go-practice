package storage

import "fmt"

type RecordNotFoundError struct {
	Value string
}

func (re *RecordNotFoundError) Error() string {
	return fmt.Sprintf("record \"%s\" not found", re.Value)
}

func RecordNotFound(value string) *RecordNotFoundError {
	return &RecordNotFoundError{value}
}
