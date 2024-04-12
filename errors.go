package smapper

import (
	"fmt"
)

var (
	ActionFieldMapping = "field mapping"
	ActionValidation   = "validation"
	ActionCallback     = "callback execution"
)

type Error struct {
	msg string
}

func (e *Error) Error() string {
	return fmt.Sprintf("smapper: %s", e.msg)
}

type FieldError struct {
	value  FieldValue
	action string
	msg    string
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("smapper: %s failed for %s.%s, %s",
		e.action,
		e.value.ParentType.Name(),
		e.value.FieldName,
		e.msg)
}
