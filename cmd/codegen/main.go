package main

import (
	"flag"
	"fmt"
	"github.com/codegen/pkg/slog"
	"os"
	"time"

	"github.com/codegen/pkg/gen"
	"github.com/codegen/pkg/output"
)

var (
	locFlag      = flag.String("c", "", "specify location of the tool's folder. Default: `cwd`")
	debugFlag    = flag.Bool("d", false, "enable debug mode")
	workersFlag  = flag.Int("w", 30, "specify number of workers available in the runtime pool")
	templateFlag = flag.Bool("ignoreTemplates", false, "ignore templates read from configuration")
)

func main() {
	defer fmt.Println()

	start := time.Now()
	flag.Parse()

	// Execute code generation.
	c := gen.Config{
		Location:         *locFlag,
		WorkerCount:      *workersFlag,
		DisableTemplates: *templateFlag,
	}
	md, mts, err := gen.Execute(c, slog.New(*debugFlag))

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
