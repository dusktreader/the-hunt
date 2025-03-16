package validator

import (
	"net/url"
	"regexp"
	"slices"
)

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func IsURL(value string) bool {
	_, err := url.ParseRequestURI(value)
	return err == nil
}

type Validator struct {
	errors map[string][]string
}

func New() *Validator {
	return &Validator{errors: make(map[string][]string)}
}

func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}

func (v *Validator) AddError(key string, message string) {
	if _, exists := v.errors[key]; !exists {
		v.errors[key] = []string{message}
	} else {
		v.errors[key] = append(v.errors[key], message)
	}
}

func (v *Validator) Errors() map[string]any {
	errors := make(map[string]any)
	for key, value := range v.errors {
		if len(value) == 1 {
			errors[key] = value[0]
		} else {
			errors[key] = value
		}
	}
	return errors
}

func (v *Validator) Check(ok bool, key string, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(uniqueValues) == len(values)
}
