package core

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
	"strings"
)

var (
	dirLikeRegex       = regexp.MustCompile("^[aA-zZ\\d/]+$")
	propTypeRegex      = regexp.MustCompile("^[aA-zZ\\d_-]+$")
	parseFileNameRegex = regexp.MustCompile("\\\\{([aA-zZ]+\\.([aA-zZ]+\\.?)+)\\\\}")
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

	// Define a custom validation tag for file names. Moreover, parses the given format into the current struct.
	_ = v.RegisterValidation("jobfilename", func(fl validator.FieldLevel) bool {
		parent := fl.Parent()
		jfn, ok := parent.Addr().Interface().(*ScopeJobFileName)
		if !ok {
			return false
		}

		matches := parseFileNameRegex.FindAllStringSubmatch(jfn.Value, -1)
		for _, val := range Map(matches, func(s []string) string { return s[1] }) {
			if !jfn.Assign(strings.Split(val, ".")) {
				return false
			}
		}
		return true
	})

	return v
}

func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}
