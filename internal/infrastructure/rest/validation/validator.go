package validation

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

func GetValidationErrorMessage(errs *validator.ValidationErrors) string {
	errsLastIndex := len(*errs) - 1
	var sb strings.Builder
	for index, err := range *errs {
		sb.WriteString(fmt.Sprintf(
			"Field '%s' failed validation: rule '%s' (value: %v)",
			err.Field(),
			err.Tag(),
			err.Value(),
		))
		if index < errsLastIndex {
			sb.WriteString("; ")
		}
	}
	return sb.String()
}
