package output

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/maxzaleski/codegen/internal"
	"github.com/maxzaleski/codegen/internal/fs"
	"github.com/maxzaleski/codegen/internal/lib"
	"github.com/maxzaleski/codegen/internal/lib/slice"
	"github.com/maxzaleski/codegen/internal/slog"
	"github.com/maxzaleski/codegen/pkg/gen/modules"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/maxzaleski/codegen/internal/core"
)

type (
	Client interface {
		PrintFinalReport(m modules.IMetrics)
		PrintError(err error)
		PrintInfo(lines ...string)
	}

	client struct {
		core.Metadata

		began          time.Time
		disableLogFile bool
		debugVerbose   bool
	}
)

// New returns a new implementation of `Client`.
func New(md core.Metadata, began time.Time, disableLogFile, debugVerbose bool) Client {
	return &client{
		Metadata: md,

		began:          began,
		disableLogFile: disableLogFile,
		debugVerbose:   debugVerbose,
	}
}

func (c *client) PrintInfo(lines ...string) {
	log.Println(infoAtom("💡", lines...))
}

func (c *client) PrintError(err error) {
	if c.debugVerbose {
		return
	}
	if !c.disableLogFile {
		defer c.writeLog(err) // writes to log file.
	}

	err, msg := lib.Unwrap(err), err.Error()
	if valErrs, ok := err.(validator.ValidationErrors); ok {
		if len(valErrs) == 1 {
			msg = err.Error()
		} else {
			// TODO: improve verbosity of problem, validation messaging
			msg = "\n"
			msg += strings.Join(
				slice.Map(valErrs, func(fe validator.FieldError) string { return "\t- " + fe.Error() }),
				"\n")
		}
	}
	log.Printf("\n%s%s\n",
		slog.Atom(slog.Red, eventPrefix("🫣"), "You've encountered an error:", msg),
		infoAtom("🐞",
			fmt.Sprintf("Please check the error log file %s for the complete stracktrace.", c.getLogDest()),
			fmt.Sprintf("If the issue persists, please do report it to me: %s 👈", slog.Atom(slog.Cyan, internal.GHIssuesURL)),
		))
}

// writeLog appends the given error to a log file in the current working directory.
//
// If the file does not exist, writeLog will create it.
func (c *client) writeLog(err1 error) {
	var err2 error
	defer func(err2 error) {
		if err2 != nil {
			log.Println(eventPrefix("💀[critical]"), "unable to write to log file")
		}
	}(err2)

	dest, stackTrace := c.getLogDest(), fmt.Sprintf("%+v", err1)

	// Check if the file already exists. If so, append the bytes.
	if fs.FileExists(dest) {
		var file *os.File
		file, err2 = os.OpenFile(dest, os.O_APPEND|os.O_WRONLY, 0644)
		if err2 != nil {
			return
		}
		defer func(file *os.File) { _ = file.Close() }(file)

		s := fmt.Sprintf("\n\nbegan=%s\n%v+", c.began, stackTrace)
		if _, err2 = file.Write([]byte(s)); err2 != nil {
			return
		}
	} else {
		// Otherwise, create the file.
		if err2 = fs.CreateFile(dest, []byte(stackTrace)); err2 != nil {
			return
		}
	}
}

func (c *client) PrintFinalReport(ms modules.IMetrics) {
	// Alphabetically sort the packages.
	jm := ms.GetJobsMetrics()
	scopes := slice.MapKeys(jm)
	sort.Strings(scopes)

	// Print metrics.go per package.
	totalFiles, seenPkgsMap := 0, make(map[string]bool)
	for _, s := range scopes {
		printScope(s)

		pms := jm[s].(map[string][]modules.MetricJob)
		for pkg := range pms {
			printPkg(pkg)

			if len(pms[pkg]) != 0 && pkg != core.UniquePkgAlias {
				if !seenPkgsMap[pkg] {
					seenPkgsMap[pkg] = true
				}
			}
			for _, mrt := range pms[pkg] {
				printFile(mrt.FileAbsolutePath, mrt.FileCreated)
				if mrt.FileCreated {
					totalFiles++
				}
			}
		}
	}

	// Print final report.
	if totalFiles == 0 {
		log.Printf("\n%s %s %s", eventPrefix("💭"), core.DomainDir, "unchanged")
		c.PrintInfo(
			"If this is unexpected, verify that a new job is correctly defined in the config file.",
			"For more information, please refer to the official documentation.",
		)
	} else {
		log.Printf("\n%s Generated %s across %s in %s.\n",
			eventPrefix("🤓"),
			slog.Atom(slog.Blue, fmt.Sprintf("%d files", totalFiles)),
			slog.Atom(slog.Blue, fmt.Sprintf("%d packages", len(slice.MapKeys(seenPkgsMap)))),
			slog.Atom(slog.Cyan, time.Since(c.began).String()),
		)
	}
}

func (c *client) getLogDest() string {
	return c.Cwd + "/codegen_error.log"
}

func printScope(name string) {
	fmt.Printf("\n🔬 %s\n", slog.Atom(slog.Bold+slog.Purple, name))
}

func printPkg(name string) {
	fmt.Printf("%s\n%s 📦 %s\n", connectorTokenNeutral, connectorToken, slog.Atom(slog.Bold+slog.Cyan, name+"/"))
}

func printFile(name string, created bool) {
	statusToken, statusColour := fileIgnoredToken, slog.Grey
	fileColour := statusColour
	if created {
		statusToken, statusColour = fileCreatedToken, slog.Green
		fileColour = slog.White
	}
	fmt.Printf("%s  %s  %s\n", connectorTokenNeutral, slog.Atom(statusColour, statusToken), slog.Atom(fileColour, name))
}

func eventPrefix(emoji string) string {
	return emoji + " " + eventToken
}

func infoAtom(emoji string, lines ...string) string {
	return fmt.Sprintf("\n%s%s %s",
		slog.Atom(slog.Grey, connectorTokenFile),
		eventPrefix(emoji),
		strings.Join(lines, "\n     "))
}
