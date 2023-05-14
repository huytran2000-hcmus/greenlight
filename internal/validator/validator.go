package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	NonFieldErrs []string
	Errors       map[string]string
}

func New() *Validator {
	return &Validator{Errors: map[string]string{}}
}

func (v *Validator) CheckError(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

func (v *Validator) IsValid() bool {
	return len(v.Errors) == 0 && len(v.NonFieldErrs) == 0
}

func (v *Validator) AddFieldError(key, message string) {
	_, exists := v.Errors[key]
	if !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrs = append(v.NonFieldErrs, message)
}

func NotBlank(val string) bool {
	return strings.TrimSpace(val) != ""
}

func LengthLessOrEqual(val string, n int) bool {
	return utf8.RuneCountInString(val) <= n
}

func PermittedValue[T comparable](val T, permittedValues ...T) bool {
	for _, permitted := range permittedValues {
		if val == permitted {
			return true
		}
	}

	return false
}

func Matches(val string, rx *regexp.Regexp) bool {
	return rx.MatchString(val)
}

func MatchesEmail(val string) bool {
	return Matches(val, EmailRX)
}

func Unique[T comparable](vals []T) bool {
	uniqueValues := map[T]struct{}{}

	for _, val := range vals {
		if _, ok := uniqueValues[val]; ok {
			return false
		}

		uniqueValues[val] = struct{}{}
	}

	return true
}
