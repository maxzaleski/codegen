package core

import (
	"github.com/go-playground/validator/v10"
	"github.com/maxzaleski/codegen/internal/lib/moddedstring"
	"reflect"
	"regexp"
	"strings"
)

var (
	dirLikeRegex  = regexp.MustCompile("(?i)^[a-z0-9/_]+$")
	propTypeRegex = regexp.MustCompile("(?i)^[a-z0-9_-]+$")
	fileNameRegex = regexp.MustCompile("(?i)^(?:[a-z0-9_-]+|\\\\\\{([a-z0-9.]+)\\\\}([a-z_-]+)?)\\.[a-z]+$")
)

// newValidator returns a new instance of `validator.Validate`.
func newValidator() *validator.Validate {
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

	// Define a custom validation tag for file names.
	_ = v.RegisterValidation("filename", func(fl validator.FieldLevel) bool {
		// ssm indexing:
		// • [0] - full match
		// • [1] - `moddedstring` substring
		ssm := fileNameRegex.FindStringSubmatch(fl.Field().String())
		if len(ssm) == 0 {
			return false
		}
		if mss := ssm[1]; mss != "" {
			moddedstring.Validate(strings.Split(mss, ".")[1:])
		}
		return true
	})

	return v
}
