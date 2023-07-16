package moddedstring

import (
	"fmt"
	"github.com/maxzaleski/codegen/internal/lib/slice"
	"regexp"
)

var stringModsRegex = regexp.MustCompile("\\\\{([aA-zZ]+\\.([aA-zZ]+\\.?)+)\\\\}")

// New attempts to parse the given modded string and returns the result.
//
// If any of the specified mods are invalid, an error is returned.
func New(src string, tokenMap map[string]string) (string, error) {
	matches_ := stringModsRegex.FindAllStringSubmatch(src, -1) // form: '{token.mod1.mod2...}'
	matches := slice.Map(matches_, func(s []string) string { return s[1] })

	ms := &moddedString{}

	// Parse all modifier instances and assign them to the modded string.
	for rKey, match := range matches {
		if ok := ms.AssignMod(rKey, match); !ok {
			return "", nil
		}
	}

	// Replace all modifier instances with a replacement key '{pkg.asCamel}' => '{n}'.
	i := len(matches)
	ms.Value = stringModsRegex.ReplaceAllStringFunc(src, func(s string) string {
		defer func(i *int) { *i-- }(&i)
		return fmt.Sprintf("{%d}", len(matches)-i)
	})

	// Applies the string modifiers.
	ms.ApplyMods(tokenMap)

	return ms.String(), nil
}
