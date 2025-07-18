package validator

import (
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New()

	return &Validator{validator: v}
}

func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}
