package core

type Spec struct {
	// Represents the global configuration.
	Global *GlobalConfig
	// Represents the packages to be generated.
	Pkgs []*Pkg
	// Represents the .codegen path.

	DirPath string
	Cwd     string
}

type GlobalConfig struct {
	Pkg *GCPkg `yaml:"pkg"`
}

type GCPkg struct {
	Extension string    `yaml:"extension"`
	Output    string    `yaml:"output"`
	Layers    []*Layer  `yaml:"layers"`
	Plugins   []*Plugin `yaml:"addons"`
}

type Plugin struct {
	Name       string               `yaml:"name"`
	Template   string               `yaml:"template"`
	FileSuffix string               `yaml:"file-suffix"`
	On         []*PluginApplication `yaml:"on,omitempty"`
}

type Layer struct {
	Name     string `yaml:"name"`
	FileName string `yaml:"file-name"`
	Template string `yaml:"template"`
}

type LayerID string

const (
	LayerIDService    LayerID = "service"
	LayerIDRepository LayerID = "repository"
)

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

func (m *Method) SortArguments() {
	m.Params = m.sortArguments(m.Params)
	m.Returns = m.sortArguments(m.Returns)
}

func (m *Method) sortArguments(args []*Argument) []*Argument {
	if len(args) == 0 {
		return args
	}

	n := len(args)
	swapped := true

	// Optimised bubble sort as we expect a small number of arguments.
	for swapped {
		swapped = false
		for i := 1; i < n; i++ {
			if args[i-1].Index > args[i].Index {
				args[i-1], args[i] = args[i], args[i-1]
				swapped = true
			}
		}
		n--
	}
	return args
}

type Interface struct {
	Description string    `yaml:"description"`
	Methods     []*Method `yaml:"methods,omitempty"`
}
