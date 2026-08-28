package components

import (
	"html/template"
	"strings"

	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
)

type FormErrorData struct {
	FieldErrors map[string][]string
	GlobalError string
	Success     bool
}

func NewFormErrorData(err *pbErrors.StandardError) *FormErrorData {
	if err == nil {
		return &FormErrorData{
			FieldErrors: make(map[string][]string),
			Success:     true,
		}
	}

	fields := pbErrors.ExpandFieldKeys(err.Fields)
	return &FormErrorData{
		FieldErrors: fields,
		GlobalError: err.GlobalError,
		Success:     false,
	}
}

func (f *FormErrorData) HasFieldError(field string) bool {
	return len(f.GetFieldErrors(field)) > 0
}

func (f *FormErrorData) GetFieldError(field string) string {
	errors := f.GetFieldErrors(field)
	if len(errors) > 0 {
		return errors[0]
	}
	return ""
}

func (f *FormErrorData) GetFieldErrors(field string) []string {
	if f == nil {
		return nil
	}
	return pbErrors.FieldMessages(f.FieldErrors, field)
}

func (f *FormErrorData) GetFieldErrorsAsString(field string) string {
	return strings.Join(f.GetFieldErrors(field), ", ")
}

func (f *FormErrorData) HasGlobalError() bool {
	return f != nil && f.GlobalError != ""
}

func (f *FormErrorData) GetFieldClasses(field string, baseClasses string) template.HTMLAttr {
	classes := baseClasses
	if f.HasFieldError(field) {
		classes += " border-red-500 focus:border-red-500 focus:ring-red-500"
	} else {
		classes += " border-gray-300 focus:border-blue-500 focus:ring-blue-500"
	}
	return template.HTMLAttr(classes)
}
