package moddedstring

import (
	"testing"
)

func TestApplyCaseModifiers(t *testing.T) {
	testCases := []struct {
		token          string
		primaryMod     CaseModifier
		secondaryMod   CaseModifier
		expectedResult string
	}{
		{
			token:          "hello_world",
			primaryMod:     CaseModifierLower,
			secondaryMod:   CaseModifierCamel,
			expectedResult: "helloWorld",
		},
		{
			token:          "hello_world",
			primaryMod:     CaseModifierUpper,
			secondaryMod:   CaseModifierSnake,
			expectedResult: "HELLO_WORLD",
		},
		{
			token:          "hello_world",
			primaryMod:     CaseModifierUpper,
			secondaryMod:   CaseModifierKebab,
			expectedResult: "HELLO-WORLD",
		},
		{
			token:          "hello_world",
			primaryMod:     CaseModifierLower,
			secondaryMod:   CaseModifierSnake,
			expectedResult: "hello_world",
		},
		{
			token:          "hello_world",
			primaryMod:     CaseModifierLower,
			secondaryMod:   CaseModifierKebab,
			expectedResult: "hello-world",
		},
		{
			token:          "hello_world",
			primaryMod:     CaseModifierNone,
			secondaryMod:   CaseModifierSnake,
			expectedResult: "hello_world",
		},
		{
			token:          "hello_world",
			primaryMod:     CaseModifierNone,
			secondaryMod:   CaseModifierKebab,
			expectedResult: "hello-world",
		},
		{
			token:          "hello_world",
			primaryMod:     CaseModifierNone,
			secondaryMod:   CaseModifierCamel,
			expectedResult: "HelloWorld",
		},
		{
			token:          "hello_world",
			primaryMod:     CaseModifierTitle,
			secondaryMod:   CaseModifierNone,
			expectedResult: "HelloWorld",
		},
	}

	for _, testCase := range testCases {
		result := applyCaseModifiers(testCase.token, testCase.primaryMod, testCase.secondaryMod)

		if result != testCase.expectedResult {
			t.Errorf("Expected '%s' but got '%s' for token: '%s' with primaryMod: '%s' and secondaryMod: '%s'",
				testCase.expectedResult, result, testCase.token, testCase.primaryMod, testCase.secondaryMod)
		}
	}
}
