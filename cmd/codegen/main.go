package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/codegen/internal/core"
	"github.com/codegen/internal/fs"
	"github.com/codegen/pkg/gen"
	"github.com/codegen/pkg/output"
	"github.com/pkg/errors"
)

var (
	locFlag   = flag.String("c", "", "location of the .codegen folder")
	debugFlag = flag.Int("d", 0, "enable debug mode")
)

func main() {
	flag.Parse()

	// Parse configuration via `.codegen` directory.
	spec, err := core.NewSpec(*locFlag)
	if err != nil {
		output.Error(err)
		os.Exit(1)
	}

	// Execute code generation.
	m, err := gen.Execute(spec, *debugFlag)
	if err != nil {
		// Create error log file.
		if err := createErrorLogFile(spec.Paths.Cwd, err); err != nil {
			output.Error(errors.Wrap(err, "failed to create error log file"))
			os.Exit(1)
		}

		output.Error(err)
		os.Exit(1)
	}

	// Output generation metrics to stdout.
	outputMetrics(m)
}

func outputMetrics(m gen.Metrics) {
	// Alphabetically sort the packages.
	keys := m.Keys()
	sort.Strings(keys)

	// Print metrics per package.
	totalGenerated := 0
	for _, pkg := range keys {
		output.Package(pkg)

		for _, file := range m.Get(pkg) {
			output.File(file.Key, file.Created)
			if file.Created {
				totalGenerated++
			}
		}
		fmt.Println()
	}

	// Print final report.
	output.Report(totalGenerated, len(keys))
	if totalGenerated == 0 {
		output.Info("If this is unexpected, verify that a new layer is correctly defined in the config file. For more information, please refer to the official documentation.")
	}
}

// createErrorLogFile creates a log file containing the error stack trace.
func createErrorLogFile(cwd string, err error) error {
	dest := cwd + "/codegen_error.log"
	bytes := []byte(fmt.Sprintf("%+v", err))
	return fs.CreateFile(dest, bytes)
}
