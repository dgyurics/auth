package model

import "fmt"

type Errors []error

func (errs Errors) Error() string {
	var errStr string
	for _, err := range errs {
		errStr += fmt.Sprintf("%s\n", err.Error())
	}
	return errStr
}
