package main

import (
	"flag"
	"github.com/maxzaleski/codegen/pkg/gen"
	"github.com/maxzaleski/codegen/pkg/output"
	"os"
	"text/template"
	"time"
)

var (
	locFlag                = flag.String("location", "", "specify location of the tool's folder; default: '{cwd}/.codegen'")
	debugFlag              = flag.Bool("debug", false, "enable debug mode; prints debug messages to stdout")
	workersFlag            = flag.Int("workers", 30, "specify number of workers available in the runtime pool")
	debugWorkerMetricsFlag = flag.Bool("workerMetrics", false, "debug must be enabled; prints worker metrics to stdout")
	deleteTmpFlag          = flag.Bool("deleteTmp", false, "deletes the dir structure at '{cwd}/tmp'")
	templateFlag           = flag.Bool("ignoreTemplates", false, "ignore templates read from configuration")
)

func init() {
	flag.Parse()
}

func main() { New(nil) }

// New instantiates the code generation process.
//
// funcMap is a map of functions that can be used in templates.
//
//	funcMap := template.FuncMap{
//		"add": func(a, b int) int {
//			return a + b
//		},
//	}
func New(funcMap template.FuncMap) {
	start := time.Now()

	// Execute code generation.
	c := gen.Config{
		DebugMode:          *debugFlag,
		DebugWorkerMetrics: *debugWorkerMetricsFlag,
		DeleteTmp:          *deleteTmpFlag,
		DisableTemplates:   *templateFlag,
		Location:           *locFlag,
		WorkerCount:        *workersFlag,

		TemplateFuncMap: funcMap,
	}
	md, mts, err := gen.Execute(c, start)

	// Instantiate output client.
	o := output.New(*md, start)

	// Handle outcome.
	if err != nil {
		o.PrintError(err)
		os.Exit(1)
	} else {
		o.PrintFinalReport(mts)
	}
}
