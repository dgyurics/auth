package model

import "fmt"

// Errors is a custom type created to enforce a specific set of values
type Errors []error

func (errs Errors) Error() string {
	var errStr string
	for _, err := range errs {
		errStr += fmt.Sprintf("%s\n", err.Error())
	}
	return errStr
}
