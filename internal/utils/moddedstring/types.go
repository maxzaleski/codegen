package moddedstring

import (
	"fmt"
	"strings"
)

type (
	moddedString struct {
		Value string
		Mods  []*stringModifier
	}

	stringModifier struct {
		ReplacementKey int
		Token          string
		Mods           []CaseModifier
	}
)

// AssignMod assigns a string modifier to the modded string.
//
// - `rKey` is the replacement key, which is the index of the string modifier in the modded string (e.g. 0 => '{0}').
//
// - `match` is the string modifier to assign (e.g. 'asCamel').
func (s *moddedString) AssignMod(rKey int, match string) bool {
	vals := strings.Split(match, ".")

	sm := &stringModifier{
		ReplacementKey: rKey, // {...asLower...} => {n}, where n is the replacement key
		Token:          vals[0],
		Mods:           make([]CaseModifier, len(vals)-1),
	}
	for _, val := range vals[1:] {
		if !PrimaryCaseModifier(val).IsValid() && !SecondaryCaseModifier(val).IsValid() {
			return false
		}
		sm.Mods = append(sm.Mods, CaseModifier(val))
	}
	s.Mods = append(s.Mods, sm)

	return true
}

func (s *moddedString) String() string {
	return s.Value
}

// ApplyMods applies the string modifiers to the string.
func (s *moddedString) ApplyMods(tokenMap map[string]string) {
	for _, mod := range s.Mods {
		token, pm, sm := mod.Token, CaseModifierNone, CaseModifierNone
		if v := tokenMap[token]; v != "" {
			token = v
		}
		for _, m := range mod.Mods {
			if PrimaryCaseModifier(m).IsValid() {
				pm = m
			} else if SecondaryCaseModifier(m).IsValid() {
				sm = m
			}
		}
		s.Value = strings.Replace(s.Value,
			/* old */ fmt.Sprintf("{%d}", mod.ReplacementKey),
			/* new */ applyCaseModifiers(token, pm, sm),
			1,
		)
	}
}
