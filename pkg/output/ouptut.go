package output

import (
	"fmt"
	"github.com/codegen/internal/fs"
	"github.com/codegen/internal/metrics"
	"github.com/codegen/internal/utils"
	"github.com/go-playground/validator/v10"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/codegen/internal"
	"github.com/codegen/internal/core"
)

type (
	Client interface {
		PrintFinalReport(m metrics.IMetrics)
		PrintError(err error)
		PrintInfo(lines ...string)
	}

	client struct {
		core.Metadata

		began time.Time
	}
)

// New returns a new implementation of `Client`.
func New(md core.Metadata, began time.Time) Client {
	return &client{md, began}
}

func (c *client) PrintInfo(lines ...string) {
	fmt.Println(infoAtom("💡", lines...))
}

func (c *client) PrintError(err error) {
	defer c.writeLog(err) // writes to log file.

	err, msg := utils.Unwrap(err), err.Error()
	if valErrs, ok := err.(validator.ValidationErrors); ok {
		if len(valErrs) == 1 {
			msg = err.Error()
		} else {
			// TODO: improve (verbosity of problem) validation messaging
			msg = "\n"
			msg += strings.Join(
				utils.Map(valErrs, func(fe validator.FieldError) string { return "\t- " + fe.Error() }),
				"\n")
		}
	}
	fmt.Printf("%s%s\n",
		atom(red, eventPrefix("🫣"), "You've encountered an error:", msg),
		infoAtom("🐞",
			fmt.Sprintf("Please check the error log file %s for the complete stracktrace.", c.getLogDest()),
			fmt.Sprintf("If the issue persists, please do report it to me: %s 👈", atom(cyan, internal.GHIssuesURL)),
		))
}

// writeLog appends the given error to a log file in the current working directory.
//
// If the file does not exist, writeLog will create it.
func (c *client) writeLog(err1 error) {
	var err2 error
	defer func(err2 error) {
		if err2 != nil {
			fmt.Println(eventPrefix("💀[critical]"), "unable to write to log file")
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

func (c *client) PrintFinalReport(m metrics.IMetrics) {
	// Alphabetically sort the packages.
	keys := m.Keys()
	sort.Strings(keys)

	// Print metrics per package.
	totalFiles, totalPkgs := 0, 0
	for _, scope := range m.Keys() {
		printScope(scope)

		pkgs := m.Get(scope)
		for pkg := range pkgs {
			printPkg(pkg)

			mrts := pkgs[pkg]
			if len(mrts) != 0 {
				totalPkgs++
			}
			for _, mrt := range pkgs[pkg] {
				printFile(mrt.FileAbsolutePath, mrt.Created)
				if mrt.Created {
					totalFiles++
				}
			}
		}
	}

	// Print final report.
	if totalFiles == 0 {
		fmt.Printf("\n%s %s %s", eventPrefix("💭"), core.DomainDir, "unchanged")
		c.PrintInfo(
			"If this is unexpected, verify that a new job is correctly defined in the config file.",
			"For more information, please refer to the official documentation.",
		)
	} else {
		fmt.Printf("\n%s Generated %s across %s in %s.\n",
			eventPrefix("🤓"),
			atom(blue, fmt.Sprintf("%d files", totalFiles)),
			atom(blue, fmt.Sprintf("%d packages", totalPkgs)),
			atom(cyan, time.Since(c.began).String()),
		)
	}
}

func (c *client) getLogDest() string {
	return c.Cwd + "/codegen_error.log"
}

func printScope(name string) {
	fmt.Printf("\n🔬 %s\n", atom(bold+purple, name))
}

func printPkg(name string) {
	fmt.Printf("%s\n%s 📦 %s\n", connectorTokenNeutral, connectorToken, atom(bold+cyan, name+"/"))
}

func printFile(name string, created bool) {
	statusToken, statusColour := fileIgnoredToken, grey
	fileColour := statusColour
	if created {
		statusToken, statusColour = fileCreatedToken, green
		fileColour = white
	}
	fmt.Printf("%s  %s  %s\n", connectorTokenNeutral, atom(statusColour, statusToken), atom(fileColour, name))
}

func infoAtom(emoji string, lines ...string) string {
	return fmt.Sprintf("\n%s%s %s",
		atom(grey, connectorTokenFile),
		eventPrefix(emoji),
		strings.Join(lines, "\n     "))
}

func atom(colour terminalColour, text ...string) string {
	return fmt.Sprintf("%s%s%s", colour, strings.Join(text, " "), white)
}

func eventPrefix(emoji string) string {
	return emoji + " " + eventToken
}
