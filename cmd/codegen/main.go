package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/codegen/internal/core"
	"github.com/codegen/pkg/gen"
	"github.com/codegen/pkg/output"
	"github.com/pkg/errors"
)

var (
	locFlag   = flag.String("c", "", "location of the .codegen folder")
	debugFlag = flag.Int("d", 0, "enable debug mode")
)

func main() {
	start := time.Now()
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
		if err := output.WriteToErrorLog(spec.Metadata.Cwd, err); err != nil {
			output.Error(errors.Wrap(err, "failed to create error log file"))
			os.Exit(1)
		}

		output.Error(err)
		os.Exit(1)
	}

	// Output generation metrics to stdout.
	outputMetrics(start, m)
}

func outputMetrics(began time.Time, m gen.Metrics) {
	// Alphabetically sort the packages.
	keys := m.Keys()
	sort.Strings(keys)

	// Print metrics per package.
	totalGenerated := 0
	for _, pkg := range keys {
		output.Package(pkg)

		for _, file := range m.Get(pkg) {
			output.File(file.Path, file.Created)
			if file.Created {
				totalGenerated++
			}
		}
		fmt.Println()
	}

	// Print final report.
	output.Report(began, totalGenerated, len(keys))
	if totalGenerated == 0 {
		output.Info(
			"If this is unexpected, verify that a new layer is correctly defined in the config file." +
				output.EventIndent("For more information, please refer to the official documentation.", true))
	}
	fmt.Println()
}
