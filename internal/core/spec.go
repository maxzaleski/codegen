package core

type Spec struct {
	// Represents the global configuration.
	GlobalConfig GlobalConfig
	// Represents the packages to be generated.
	Pkgs []Pkg
}

type GlobalConfig struct {
	// Represents the global configuration.
	General *GCGeneral `yaml:"general"`
	// Represents the global `pkg` configuration.
	Pkg *GCPkg `yaml:"pkg"`
	// Represents the global addon configuration.
	Addons []Addon `yaml:"addons"`
}

type GCGeneral struct {
	CommentPrefix string `yaml:"comment-prefix"`
}

type GCPkg struct {
	BasePath        string  `yaml:"base-path"`
	ServiceLayer    *Layer  `yaml:"service,omitempty"`
	RepositoryLayer *Layer  `yaml:"repository,omitempty"`
	Addons          []Addon `yaml:"addons"`
}

type Addon struct {
	Name       string             `yaml:"name"`
	Template   string             `yaml:"template"`
	FileSuffix string             `yaml:"file-suffix"`
	On         []AddonApplication `yaml:"on,omitempty"`
}

type Layer struct {
	FileName      string         `yaml:"file-name"`
	AlwaysInclude *AlwaysInclude `yaml:"always-include,omitempty"`
}

type AlwaysInclude struct {
	Params  []Argument       `yaml:"params,omitempty"`
	Returns []ReturnArgument `yaml:"returns,omitempty"`
}

type Argument struct {
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
	Index int8   `yaml:"index"`
}

type ReturnArgument = Argument

type AddonApplication string

const (
	AddonApplicationService    AddonApplication = "service"
	AddonApplicationRepository AddonApplication = "repository"
)

type Entity struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type EntityWithScope struct {
	Entity
	Scope EntityScope `yaml:"scope"`
}

type EntityScope string

const (
	EntityScopePublic    EntityScope = "public"
	EntityScopePrivate   EntityScope = "private"
	EntityScopeProtected EntityScope = "protected"
)

type Pkg struct {
	Entity
	Models    []Model   `yaml:"models,omitempty"`
	Interface Interface `yaml:"interface"`
}

type Model struct {
	EntityWithScope
	Extends    string     `yaml:"extends"`
	Implements string     `yaml:"implements"`
	Properties []Property `yaml:"props,omitempty"`
	Methods    []Method   `yaml:"methods,omitempty"`
}

type Property struct {
	EntityWithScope
	Type   string                  `yaml:"type"`
	Addons *map[string]interface{} `yaml:"addons,omitempty"`
}

type Method struct {
	EntityWithScope
	Params  []Argument       `yaml:"params,omitempty"`
	Returns []ReturnArgument `yaml:"returns,omitempty"`
}

type Interface struct {
	Description string   `yaml:"description"`
	Methods     []Method `yaml:"methods,omitempty"`
}
