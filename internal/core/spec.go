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
		Config: &Config{},
		Pkgs:   make([]*Package, 0),
		Metadata: &Metadata{
			PkgsLastModifiedMap: make(map[string]int64),
		},
	}
}

// Metadata represents the location metadata for the current generation.
type Metadata struct {
	// Location of the '.codegen' directory.
	CodegenDir string
	// Location of the current working directory.
	Cwd string
	// Represents the last modified time of the packages (values in unix).
	PkgsLastModifiedMap map[string]int64
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
		Key       string             `yaml:"key" validate:"required"`
		FileName  string             `yaml:"file-name" validate:"required,filename"`
		Templates []ScopeJobTemplate `yaml:"templates" validate:"required,dive"`
		// Excludes is a list of packages that do not need to be considered for the current job.
		Excludes []string `yaml:"exclude" validate:"omitempty,alpha"`
		// Override is a flag that indicates whether the current job should override an existing file.
		Override bool `yaml:"override" validate:"boolean"`
		// OverrideOn indicates whether the job should override an existing file based on provided conditions.
		OverrideOn map[string]ScopeJobOverride `yaml:"override-on" validate:"omitempty,dive"`
		Unique     bool                        `yaml:"unique" validate:"boolean"`
	}

	ScopeJobOverride struct {
		Model     bool `yaml:"model"`
		Interface bool `yaml:"interface"`
	}

	ScopeJobTemplate struct {
		Primary bool   `yaml:"primary" validate:"boolean"`
		Name    string `yaml:"name" validate:"required"`
	}
)

// Copy deep copies the struct instance.
func (s *ScopeJob) Copy() *ScopeJob {
	sCopy := *s
	copy(sCopy.Templates, s.Templates)
	return &sCopy
}

func (o *ScopeJobOverride) Merge(newO ScopeJobOverride) {
	if !o.Model && newO.Model {
		o.Model = true
	}
	if !o.Interface && newO.Interface {
		o.Interface = true
	}
}

func (o *ScopeJobOverride) AsSlice() []bool {
	return []bool{ // order matters.
		o.Model,
		o.Interface,
	}
}

func (o *ScopeJobOverride) Set(pi int, val bool) {
	switch pi {
	case 0:
		o.Model = val
	case 1:
		o.Interface = val
	}
}

type (
	// Domain represents a generic top-level self-contained domain of operations.
	Domain struct {
		Scopes []*DomainScope `yaml:"scopes" validate:"dive"`
	}

	// DomainScope represents a generic scope under a domain.
	DomainScope struct {
		Key        string      `yaml:"key" validate:"required"`
		Output     string      `yaml:"output" validate:"required,dirlike"`
		Inline     bool        `yaml:"inline" validate:"boolean"`
		Jobs       []*ScopeJob `yaml:"jobs" validate:"dive"`
		ParentType DomainType  `yaml:"-"`
	}

	DomainType string
)

const (
	DomainTypeHttp DomainType = "domain_http"
	DomainTypePkg  DomainType = "domain_pkg"
)

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
	sort.Slice(s, func(i, j int) bool { return s[i].Index > s[j].Index })
}

// Interface represents a generic interface definition.
type Interface struct {
	Description string      `yaml:"description"`
	Methods     []*Function `yaml:"methods,omitempty" validate:"dive"`
}
