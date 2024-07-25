package errors

import "fmt"

type Expected interface {
	IsExpected() bool
	error
}

func Unexpected(text string) error {
	return &Error{
		Message:  text,
		Expected: false,
	}
}

func Unexpectedf(format string, a ...any) error {
	return &Error{
		Message:  fmt.Sprintf(format, a...),
		Expected: false,
	}
}

func New(text string) error {
	return &Error{
		Message:  text,
		Expected: true,
	}
}

func Newf(format string, a ...any) error {
	return &Error{
		Message:  fmt.Sprintf(format, a...),
		Expected: true,
	}
}

type Error struct {
	Message  string
	Expected bool
}

func (e *Error) IsExpected() bool {
	return e.Expected
}

func (e *Error) Error() string {
	return e.Message
}
