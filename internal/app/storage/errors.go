package storage

import "fmt"

type RecordNotFoundError struct {
	Value string
}

func (re *RecordNotFoundError) Error() string {
	return fmt.Sprintf("record \"%s\" not found", re.Value)
}

func NewRecordNotFoundError(value string) *RecordNotFoundError {
	return &RecordNotFoundError{value}
}

type RecordConflictError struct {
	OldRecord Record
}

func (re *RecordConflictError) Error() string {
	return fmt.Sprintf("record already exists with short: \"%s\"", re.OldRecord.Short)
}

func NewRecordConflictError(record Record) *RecordConflictError {
	return &RecordConflictError{record}
}
