package core

type Spec struct {
	// Represents the global configuration.
	GlobalConfig *GlobalConfig
	// Represents the packages to be generated.
	Pkgs []*Pkg
}

type GlobalConfig struct {
	CommentPrefix string `yaml:"comment-prefix"`
	Pkg           *GCPkg `yaml:"pkg"`
}

type GCPkg struct {
	Extension       string    `yaml:"extension"`
	BasePath        string    `yaml:"base-path"`
	ServiceLayer    *Layer    `yaml:"service,omitempty"`
	RepositoryLayer *Layer    `yaml:"repository,omitempty"`
	Plugins         []*Plugin `yaml:"addons"`
}

type Plugin struct {
	Name       string               `yaml:"name"`
	Template   string               `yaml:"template"`
	FileSuffix string               `yaml:"file-suffix"`
	On         []*PluginApplication `yaml:"on,omitempty"`
}

type Layer struct {
	FileName      string         `yaml:"file-name"`
	Template      string         `yaml:"template"`
	AlwaysInclude *AlwaysInclude `yaml:"always-include,omitempty"`
}

type AlwaysInclude struct {
	Params  []*Argument       `yaml:"params,omitempty"`
	Returns []*ReturnArgument `yaml:"returns,omitempty"`
}

type Argument struct {
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
	Index int8   `yaml:"index"`
}

type ReturnArgument = Argument

type PluginApplication string

const (
	AddonApplicationService    PluginApplication = "service"
	AddonApplicationRepository PluginApplication = "repository"
)

type Entity struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type EntityWithScope struct {
	Entity `yaml:",inline"`
	Scope  EntityScope `yaml:"scope"`
}

type EntityScope string

const (
	EntityScopePublic    EntityScope = "public"
	EntityScopePrivate   EntityScope = "private"
	EntityScopeProtected EntityScope = "protected"
)

type Pkg struct {
	Entity    `yaml:",inline"`
	Extension string     `yaml:"extension"`
	Models    []Model    `yaml:"models,omitempty"`
	Interface *Interface `yaml:"interface"`
}

type Model struct {
	EntityWithScope `yaml:",inline"`
	Extends         string     `yaml:"extends"`
	Implements      string     `yaml:"implements"`
	Properties      []Property `yaml:"props,omitempty"`
	Methods         []Method   `yaml:"methods,omitempty"`
}

type Property struct {
	EntityWithScope `yaml:",inline"`
	Type            string                  `yaml:"type"`
	Addons          *map[string]interface{} `yaml:"addons,omitempty"`
}

type Method struct {
	EntityWithScope `yaml:",inline"`
	Params          []*Argument       `yaml:"params,omitempty"`
	Returns         []*ReturnArgument `yaml:"returns,omitempty"`
}

type Interface struct {
	Description string    `yaml:"description"`
	Methods     []*Method `yaml:"methods,omitempty"`
}
