// Copyright 2019 Anapaya Systems

package metrics

import "fmt"

var _ error = (*Error)(nil)

type Error struct {
	label string
	err   error
}

func GetErrorLabel(e error) string {
	if e, _ := e.(*Error); e != nil {
		return e.label
	}
	panic(fmt.Sprintf("Error does not contain label. err=%v", e))
}

func NewError(value string, err error) error {
	return &Error{
		label: value,
		err:   err,
	}
}

func (e *Error) Error() string {
	return e.err.Error()
}
