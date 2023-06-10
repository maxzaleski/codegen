package main

import (
	"flag"
	"fmt"
	"github.com/codegen/pkg/slog"
	"os"
	"time"

	"github.com/codegen/pkg/gen"
	"github.com/codegen/pkg/output"
	"github.com/pkg/errors"
)

var (
	locFlag     = flag.String("c", "", "location of the .codegen folder")
	debugFlag   = flag.Bool("d", false, "enable debug mode")
	workersFlag = flag.Int("w", 30, "specify worker count")
)

func main() {
	defer fmt.Println() // Creates a final space after the output.

	start := time.Now()
	flag.Parse()

	// Execute code generation.
	md, m, err := gen.Execute(*locFlag, *workersFlag, slog.New(*debugFlag))
	if err != nil {
		output.Error(err)
		if err := output.ErrorFile(md.Cwd, err); err != nil {
			output.Error(errors.Wrap(err, "failed to create error log file"))
		}
		os.Exit(1)
	}

	// Output execution metrics.
	output.Metrics(start, m)
}
