package moddedstring

// Validate validates the given modifier tokens.
//
//	mods := ["asLower", "fooBar"]
//	mods[1] => valid
//	mods[2] => invalid
func Validate(mods []string) bool {
	for _, t := range mods {
		if !CaseModifier(t).IsValid() {
			return false
		}
	}
	return true
}
