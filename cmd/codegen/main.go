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
	debugVerboseFlag       = flag.Bool("debugVerbose", false, "enable debug verbose mode; prints verbose error messages to stdout")
	workersFlag            = flag.Int("workers", 30, "specify number of workers available in the runtime concierge")
	debugWorkerMetricsFlag = flag.Bool("workerMetrics", false, "debug must be enabled; prints worker metrics.go to stdout")
	deleteTmpFlag          = flag.Bool("deleteTmp", false, "deletes the dir structure at '{cwd}/tmp'")
	ignoreTemplatesFlag    = flag.Bool("ignoreTemplates", false, "ignore templates read from configuration")
	disableLogFileFlag     = flag.Bool("disableLogFile", false, "ignore templates read from configuration")
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
		DebugVerbose:       *debugVerboseFlag,
		DebugWorkerMetrics: *debugWorkerMetricsFlag,
		DeleteTmp:          *deleteTmpFlag,
		IgnoreTemplates:    *ignoreTemplatesFlag,
		DisableLogFile:     *disableLogFileFlag,
		Location:           *locFlag,
		WorkerCount:        *workersFlag,

		TemplateFuncMap: funcMap,
	}
	res, err := gen.Execute(c, start)

	// Instantiate output client.
	o := output.New(*res.Metadata, start, c.DisableLogFile, c.DebugVerbose)

	// Handle outcome.
	if err != nil {
		o.PrintError(err)
		os.Exit(1)
	} else {
		o.PrintFinalReport(res.Metrics)
	}
}
