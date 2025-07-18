package validator

import (
	play "github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	V *play.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.V.Struct(i)
}
