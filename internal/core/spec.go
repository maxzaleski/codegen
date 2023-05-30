package core

import "sort"

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

// GetOutPath returns the absolute path of the final output directory.
func (s *Spec) GetOutPath() string {
	return s.Metadata.Cwd + "/" + s.Config.Output
}

// Metadata represents the location metadata for the current generation.
type Metadata struct {
	// Location of the '.codegen' directory.
	DomainDir string
	// Location of the current working directory.
	Cwd string
}

// Config represents the configuration for the current generation.
type Config struct {
	Lang   string           `yaml:"lang" validate:"lowercase,alpha,lt=5"`
	Output string           `yaml:"output" validate:"required,dirlike"`
	Jobs   []*GenerationJob `yaml:"jobs" validate:"dive"`
}

// GenerationJob represents a job to be performed for the current generation.
//
// Each package runs through the entire list of jobs. To opt out, specify the package name in the `exclude` field.
type GenerationJob struct {
	Key           string     `yaml:"name" validate:"required,alpha"`
	Destination   string     `yaml:"destination" validate:"required,dirlike"`
	Template      string     `yaml:"template" validate:"required"`
	Lang          string     `yaml:"lang"  validate:"lowercase,alpha"`
	DisableEmbeds bool       `yaml:"disable-embeds" validate:"boolean"`
	Exclude       Exclusions `yaml:"exclude" validate:"omitempty,alpha"`
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
