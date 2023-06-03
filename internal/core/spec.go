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
	PkgDomain = Domain[*PkgDomainScope]

	// PkgDomainScope represents a scope under the 'pkg' domain.
	PkgDomainScope = DomainScope[PkgScopeJob]

	// PkgScopeJob represents a job to be performed under the current 'Pkg' scope.
	PkgScopeJob = *ScopeJob
)

type (
	HttpDomain = Domain[*HttpScope]

	// HttpScope represents a scope under the 'Http' domain.
	HttpScope = DomainScope[HttpScopeJob]

	// HttpScopeJob represents a job to be performed under the current 'Http' scope.
	HttpScopeJob struct {
		*ScopeJob `yaml:",inline"`
		Concat    bool `yaml:"concat" validate:"boolean"`
	}
)

type (
	ScopeJobPresence interface {
		Get() *ScopeJob
	}

	// ScopeJob represents a generic job to be performed under the current scope.
	ScopeJob struct {
		Key           string            `yaml:"key" validate:"required"`
		FileName      *ScopeJobFileName `yaml:",inline" validate:"dive"`
		Template      string            `yaml:"template" validate:"required"`
		DisableEmbeds bool              `yaml:"disable-embeds" validate:"boolean"`
		Excludes      Exclusions        `yaml:"exclude" validate:"omitempty,alpha"`
	}

	ScopeJobFileName struct {
		Value string `yaml:"file-name" validate:"required,jobfilename"`

		Token     string
		Modifiers []CaseModifier
	}
)

func (s *ScopeJob) Get() *ScopeJob {
	return s
}

func (jfn *ScopeJobFileName) Assign(vals []string) bool {
	if jfn.Modifiers == nil {
		jfn.Modifiers = make([]CaseModifier, 0, len(vals)-1)
	}

	jfn.Token = vals[0]

	for _, val := range vals[1:] {
		isPrimary, isSecondary := PrimaryCaseModifier(val).IsValid(), SecondaryCaseModifier(val).IsValid()
		if !isPrimary && !isSecondary {
			return false
		}
		jfn.Modifiers = append(jfn.Modifiers, CaseModifier(val))
	}

	return true
}

type (
	// Domain represents a generic top-level self-contained domain of operations.
	Domain[S any] struct {
		Scopes []S `yaml:"scopes" validate:"dive"`
	}

	// DomainScope represents a generic scope under a domain.
	DomainScope[J ScopeJobPresence] struct {
		Key    string `yaml:"key" validate:"required"`
		Output string `yaml:"output" validate:"required,dirlike"`
		Jobs   []J    `yaml:"jobs" validate:"dive"`
	}
)

const (
	CaseModifierNone  CaseModifier = ""
	CaseModifierLower CaseModifier = "asLower"
	CaseModifierUpper CaseModifier = "asUpper"
	CaseModifierTitle CaseModifier = "asTitle"

	CaseModifierSnake CaseModifier = "asSnake"
	CaseModifierCamel CaseModifier = "asCamel"
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
