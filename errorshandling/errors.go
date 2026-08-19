package errorshandling

import (
	"bytes"
	"fmt"
)


type Error string

// Error implements the standard Go error interface.
func (e Error) Error() string {
	return string(e)
}

// Declare the string error enum constants.
const (
	ErrNotFound       Error = "resource not found"
	ErrDataWasntProvided       Error = "resource wasn't provided"
	ErrInternal       Error = "Internal error with ther resource "
	ErrInvalidData       Error = "Invalid data"
	ErrCreatingTheProcess       Error = "Can't setup the porcess"
	ErrTaskMaster       Error = "taskmaster error"
	ErrInvalidCMD     Error = "invalid cmd"
)


type ErrorReporter struct {
	v any
	err Error 
}

func NewErrorReporter(err Error, v any) *ErrorReporter{
	return &ErrorReporter{v : v, err :  err}
}


func (e *ErrorReporter) String() string {
	return fmt.Sprintf("%s : %v", e.err, e.v)
}


func (e *ErrorReporter) Report(b *bytes.Buffer) {
	fmt.Fprintf(b, "%v\n", e)
}