package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/codegen/pkg/gen"
	"github.com/codegen/pkg/output"
)

var (
	locFlag      = flag.String("location", "", "specify location of the tool's folder. Default: `cwd`")
	debugFlag    = flag.Bool("debug", false, "enable debug mode")
	workersFlag  = flag.Int("workers", 30, "specify number of workers available in the runtime pool")
	templateFlag = flag.Bool("ignoreTemplates", false, "ignore templates read from configuration")
)

func init() {
	flag.Parse()
}

func main() {
	defer fmt.Println()

	start := time.Now()

	// Execute code generation.
	c := gen.Config{
		Location:         *locFlag,
		WorkerCount:      *workersFlag,
		DisableTemplates: *templateFlag,
		DebugMode:        *debugFlag,
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
