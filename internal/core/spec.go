package core

import (
	"sort"
)

// Spec represents the specification for the current generation.
type Spec struct {
	// Represents the package configuration.
	Config *Config
	// Represents the packages to be generated.
	Pkgs []*Package `validate:"dive"`
	// Represents the metadata of the current generation.
	Metadata *Metadata `validate:"required"`
}

func newSpec() *Spec {
	return &Spec{
		Config:   &Config{},
		Pkgs:     make([]*Package, 0),
		Metadata: &Metadata{},
	}
}

// Metadata represents the location metadata for the current generation.
type Metadata struct {
	// Location of the '.codegen' directory.
	CodegenDir string
	// Location of the current working directory.
	Cwd string
}

// Config represents the configuration for the current generation.
type Config struct {
	PkgDomain  *PkgDomain  `yaml:"pkg" validate:"dive"`
	HttpDomain *HttpDomain `yaml:"http" validate:"dive"`
}

type (
	PkgDomain = Domain

	// PkgDomainScope represents a scope under the 'pkg' domain.
	PkgDomainScope = DomainScope
)

type (
	HttpDomain = Domain

	// HttpDomainScope represents a scope under the 'Http' domain.
	HttpDomainScope = DomainScope
)

type (
	// ScopeJob represents a generic job to be performed under the current scope.
	ScopeJob struct {
		Key              string             `yaml:"key" validate:"required"`
		FileName         *ScopeJobFileName  `yaml:",inline" validate:"dive"`
		Templates        []ScopeJobTemplate `yaml:"templates" validate:"required,dive"`
		DisableEmbeds    bool               `yaml:"disable-templates" validate:"boolean"`
		Excludes         Exclusions         `yaml:"exclude" validate:"omitempty,alpha"`
		Unique           bool               `yaml:"unique" validate:"boolean"`
		FileAbsolutePath string             `yaml:"-"`
	}

	ScopeJobTemplate struct {
		Primary bool   `yaml:"primary" validate:"boolean"`
		Name    string `yaml:"name" validate:"required"`
	}

	ScopeJobFileName struct {
		Value     string `yaml:"file-name" validate:"required,jobfilename"`
		Extension string `yaml:"-"`
		Mods      []*FileNameMod
	}

	FileNameMod struct {
		Key       int
		Token     string
		Modifiers []CaseModifier
	}
)

// Copy deep copies the struct instance.
func (s *ScopeJob) Copy() *ScopeJob {
	sCopy, fnCopy := *s, *s.FileName
	sCopy.FileName = &fnCopy
	copy(sCopy.Templates, s.Templates)
	return &sCopy
}

const modPkgReplacementToken = "pkg"

func (jfn *ScopeJobFileName) Assign(key int, vals []string) bool {
	mod := &FileNameMod{
		Key:       key,
		Token:     vals[0],
		Modifiers: make([]CaseModifier, 0, len(vals)-1),
	}
	for _, val := range vals[1:] {
		isPrimary, isSecondary := PrimaryCaseModifier(val).IsValid(), SecondaryCaseModifier(val).IsValid()
		if !isPrimary && !isSecondary {
			return false
		}
		mod.Modifiers = append(mod.Modifiers, CaseModifier(val))
	}
	jfn.Mods = append(jfn.Mods, mod)

	return true
}

type (
	// Domain represents a generic top-level self-contained domain of operations.
	Domain struct {
		Scopes []*DomainScope `yaml:"scopes" validate:"dive"`
	}

	// DomainScope represents a generic scope under a domain.
	DomainScope struct {
		Key            string      `yaml:"key" validate:"required"`
		Output         string      `yaml:"output" validate:"required,dirlike"`
		Inline         bool        `yaml:"inline" validate:"boolean"`
		AbsoluteOutput string      `yaml:"-"`
		Jobs           []*ScopeJob `yaml:"jobs" validate:"dive"`
	}

	DomainType string
)

const (
	DomainTypeHttp DomainType = "domain_http"
	DomainTypePkg  DomainType = "domain_pkg"
)

const (
	CaseModifierNone  CaseModifier = ""
	CaseModifierLower CaseModifier = "asLower"
	CaseModifierUpper CaseModifier = "asUpper"
	CaseModifierTitle CaseModifier = "asTitle"

	CaseModifierSnake CaseModifier = "asSnake"
	CaseModifierCamel CaseModifier = "asCamel"
	CaseModifierKebab CaseModifier = "asKebab"
)

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
		CaseModifierCamel:
		return true
	default:
		return false
	}
}

type Exclusions []string

func (e Exclusions) Contains(val string) bool {
	for _, el := range e {
		if el == val {
			return true
		}
	}
	return false
}

// FnParameter represents a function argument.
type FnParameter struct {
	Name  string `yaml:"name" validate:"required,alphanum"`
	Type  string `yaml:"type" validate:"required,alpha"`
	Index int8   `yaml:"index" validate:"required,number,gt=-1"`
}

// ReturnParameter represents a function's return parameter.
//
// Applicable to languages that support multiple values as return arguments.
type ReturnParameter = FnParameter

type Entity struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type EntityWithScope struct {
	Entity `yaml:",inline"`
	Scope  EntityScope `yaml:"scope" validate:"enum:EntityScope"`
}

type EntityScope string

func (es EntityScope) IsValid() bool {
	switch es {
	case EntityScopeProtected,
		EntityScopePrivate,
		EntityScopePublic:
		return true
	default:
		return false
	}
}

const (
	EntityScopePublic    EntityScope = "public"
	EntityScopePrivate   EntityScope = "private"
	EntityScopeProtected EntityScope = "protected"
)

const UniquePkgAlias = "[unique]"

type Package struct {
	Entity    `yaml:",inline"`
	Models    []Model    `yaml:"models,omitempty" validate:"dive"`
	Interface *Interface `yaml:"interface" validate:"dive"`
}

// Model represents a generic domain model.
type Model struct {
	EntityWithScope `yaml:",inline"`
	Extends         string          `yaml:"extends"`
	Implements      string          `yaml:"implements"`
	Properties      []ModelProperty `yaml:"props,omitempty" validate:"dive"`
	Methods         []Function      `yaml:"methods,omitempty" validate:"dive"`
}

// ModelProperty represents a generic property definition.
type ModelProperty struct {
	EntityWithScope `yaml:",inline" validate:"dive"`
	Type            string                  `yaml:"type" validate:"required,proptype"`
	Addons          *map[string]interface{} `yaml:"addons,omitempty"`
}

// Function represents a generic function definition.
type Function struct {
	EntityWithScope `yaml:",inline"`
	Params          []*FnParameter     `yaml:"params,omitempty"`
	Returns         []*ReturnParameter `yaml:"returns,omitempty"`
}

func (m *Function) SortParams() {
	m.sort(m.Params)
	m.sort(m.Returns)
}

func (m *Function) sort(s []*FnParameter) {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Index > s[j].Index
	})
}

// Interface represents a generic interface definition.
type Interface struct {
	Description string      `yaml:"description"`
	Methods     []*Function `yaml:"methods,omitempty" validate:"dive"`
}
