package smapper

import (
	"fmt"
)

type Error struct {
	msg string
}

func (e *Error) Error() string {
	return fmt.Sprintf("smapper: %s", e.msg)
}

type FieldError struct {
	value FieldValue
	msg   string
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("smapper: field mapping failed for %s.%s, %s",
		e.value.ParentType.Name(),
		e.value.FieldName,
		e.msg)
}

type ValidationError struct {
	value         FieldValue
	validatorName string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("smapper: validator %s failed for %s.%s",
		e.validatorName,
		e.value.ParentType.Name(),
		e.value.FieldName)
}

type CallbackError struct {
	value FieldValue
	msg   string
}

func (e *CallbackError) Error() string {
	return fmt.Sprintf("smapper: callback execution failed for %s.%s, %s",
		e.value.ParentType.Name(),
		e.value.FieldName,
		e.msg)
}
