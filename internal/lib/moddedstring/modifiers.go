package moddedstring

import "github.com/iancoleman/strcase"

const (
	CaseModifierNone  CaseModifier = ""
	CaseModifierLower CaseModifier = "asLower"
	CaseModifierUpper CaseModifier = "asUpper"
	CaseModifierTitle CaseModifier = "asTitle"

	CaseModifierSnake CaseModifier = "asSnake"
	CaseModifierCamel CaseModifier = "asCamel"
	CaseModifierKebab CaseModifier = "asKebab"
)

func applyCaseModifiers(token string, pm CaseModifier, sm CaseModifier) (result string) {
	if pm == CaseModifierLower && sm == CaseModifierCamel {
		result = strcase.ToLowerCamel(token) // helloWorld
	} else if pm == CaseModifierUpper && sm == CaseModifierSnake {
		result = strcase.ToScreamingSnake(token) // HELLO_WORLD
	} else if pm == CaseModifierUpper && sm == CaseModifierKebab {
		result = strcase.ToScreamingKebab(token) // HELLO-WORLD
	} else if pm == CaseModifierLower && sm == CaseModifierSnake ||
		pm == CaseModifierNone && sm == CaseModifierSnake {
		result = strcase.ToSnake(token) // hello_world
	} else if pm == CaseModifierLower && sm == CaseModifierKebab ||
		pm == CaseModifierNone && sm == CaseModifierKebab {
		result = strcase.ToKebab(token) // hello-world
	} else if pm == CaseModifierTitle ||
		pm == CaseModifierNone && sm == CaseModifierCamel {
		result = strcase.ToCamel(token) // HelloWorld
	} else {
		result = token
	}
	return
}

type CaseModifier string

func (m CaseModifier) IsValid() bool {
	return PrimaryCaseModifier(m).IsValid() ||
		SecondaryCaseModifier(m).IsValid()
}

type PrimaryCaseModifier CaseModifier

func (m PrimaryCaseModifier) IsValid() bool {
	switch CaseModifier(m) {
	case CaseModifierNone,
		CaseModifierLower,
		CaseModifierUpper,
		CaseModifierTitle:
		return true
	default:
		return false
	}
}

type SecondaryCaseModifier CaseModifier

func (m SecondaryCaseModifier) IsValid() bool {
	switch CaseModifier(m) {
	case CaseModifierNone,
		CaseModifierSnake,
		CaseModifierCamel,
		CaseModifierKebab:
		return true
	default:
		return false
	}
}
