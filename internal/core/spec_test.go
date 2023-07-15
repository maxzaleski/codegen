package core

import (
	"testing"
)

func TestScopeJobFileName_Assign(t *testing.T) {
	//tests := []struct {
	//	name       string
	//	vals       []string
	//	want       bool
	//	wantStruct *ScopeJobFileName
	//}{
	//	{
	//		name: "valid_single_primary",
	//		vals: []string{"test", string(CaseModifierLower)},
	//		want: true,
	//		wantStruct: &ScopeJobFileName{
	//			Token:     "test",
	//			Mods: []CaseModifier{CaseModifierLower},
	//		},
	//	},
	//	{
	//		name: "valid_single_secondary",
	//		vals: []string{"test", string(CaseModifierSnake)},
	//		want: true,
	//		wantStruct: &ScopeJobFileName{
	//			Token:     "test",
	//			Mods: []CaseModifier{CaseModifierSnake},
	//		},
	//	},
	//	{
	//		name: "valid_multiple",
	//		vals: []string{"test", string(CaseModifierUpper), string(CaseModifierSnake)},
	//		want: true,
	//		wantStruct: &ScopeJobFileName{
	//			Token:     "test",
	//			Mods: []CaseModifier{CaseModifierUpper, CaseModifierSnake},
	//		},
	//	},
	//	{
	//		name: "invalid_single",
	//		vals: []string{"test", "foobar"},
	//		want: false,
	//		wantStruct: &ScopeJobFileName{
	//			Token:     "test",
	//			Mods: []CaseModifier{},
	//		},
	//	},
	//	{
	//		name: "invalid_multiple",
	//		vals: []string{"test", string(CaseModifierSnake), string(CaseModifierUpper), "foobar"},
	//		want: false,
	//		wantStruct: &ScopeJobFileName{
	//			Token:     "test",
	//			Mods: []CaseModifier{CaseModifierSnake, CaseModifierUpper},
	//		},
	//	},
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		s := &ScopeJobFileName{}
	//		if ok := s.AssignMod(tt.vals); ok != tt.want {
	//			t.Errorf("AssignMod() = %v, want %v", ok, tt.want)
	//		}
	//		assert.Equal(t, s.Mods, tt.wantStruct.Mods)
	//	})
	//}
}
