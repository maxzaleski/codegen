package output

import (
	"fmt"
	"github.com/codegen/internal/metrics"
	"sort"
	"time"
)

func Metrics(began time.Time, m metrics.IMetrics) {
	fmt.Println()

	// Alphabetically sort the packages.
	keys := m.Keys()
	sort.Strings(keys)

	// Print metrics per package.
	totalGenerated := 0
	for _, pkg := range keys {
		Package(pkg)

		for _, file := range m.Get(pkg) {
			File(file.FileAbsolutePath, file.Created)
			if file.Created {
				totalGenerated++
			}
		}
		fmt.Println()
	}

	// Print final report.
	Report(began, totalGenerated, len(keys))
	if totalGenerated == 0 {
		Info(
			"If this is unexpected, verify that a new job is correctly defined in the config file.",
			"For more information, please refer to the official documentation.",
		)
	}
}
