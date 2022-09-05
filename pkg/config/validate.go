package config

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(input interface{}) error {
	return validate.Struct(input)
}
