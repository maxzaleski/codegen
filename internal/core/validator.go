package core

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
	"strings"
)

var (
	dirLikeRegex  = regexp.MustCompile("^[aA-zZ\\d/]+$")
	propTypeRegex = regexp.MustCompile("^[aA-zZ\\d_-]+$")
)

// NewValidator returns a new instance of `validator.Validate`.
func NewValidator() *validator.Validate {
	v := validator.New()

	// Remaps `validator.FieldError.Field()` to return the "yaml" structure tag args.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tag := fld.Tag.Get("yaml")
		if tag == "" {
			tag = fld.Tag.Get("param")
		}
		name := strings.SplitN(tag, ",", 2)[0]
		if name == "-" {
			name = ""
		}
		return name
	})

	// Define a custom validation tag for enum values.
	_ = v.RegisterValidation("enum", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()

		switch fl.Param() {
		case "EntityScope":
			return EntityScope(val).IsValid()
		default:
			return false
		}
	})

	// Define a custom validation tag for directory-like values as we only care to produce a valid directory path, not
	// establishing whether the directory exists (see: 'dirpath').
	_ = v.RegisterValidation("dirlike", func(fl validator.FieldLevel) bool {
		return dirLikeRegex.MatchString(fl.Field().String())
	})

	// Define a custom validation tag for property type values.
	_ = v.RegisterValidation("proptype", func(fl validator.FieldLevel) bool {
		return propTypeRegex.MatchString(fl.Field().String())
	})

	return v
}
