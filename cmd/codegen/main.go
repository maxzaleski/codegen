package main

import (
	"flag"
	"fmt"
	"github.com/codegen/pkg/slog"
	"os"
	"sort"
	"time"

	"github.com/codegen/pkg/gen"
	"github.com/codegen/pkg/output"
	"github.com/pkg/errors"
)

var (
	locFlag   = flag.String("c", "", "location of the .codegen folder")
	debugFlag = flag.Bool("d", false, "enable debug mode")
)

func main() {
	// Creates space after the output.
	defer fmt.Println()

	start := time.Now()
	flag.Parse()

	// Execute code generation.
	md, m, err := gen.Execute(*locFlag, slog.New(*debugFlag))
	if err != nil {
		output.Error(err)
		if err := output.WriteToErrorLog(md.Cwd, err); err != nil {
			output.Error(errors.Wrap(err, "failed to create error log file"))
		}
		os.Exit(1)
	}

	// Output generation metrics to stdout.
	outputMetrics(start, m)
}

func outputMetrics(began time.Time, m gen.Metrics) {
	fmt.Println()

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
			"If this is unexpected, verify that a new job is correctly defined in the config file.",
			"For more information, please refer to the official documentation.",
		)
	}
}
