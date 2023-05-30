package output

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"regexp"
	"strings"
	"time"

	"github.com/codegen/internal"
	"github.com/codegen/internal/core"
)

// terminalColour represents a terminal colour.
type terminalColour string

const (
	grey   terminalColour = "\033[90m"
	white  terminalColour = "\033[0m"
	green  terminalColour = "\033[32m"
	red    terminalColour = "\033[31m"
	purple terminalColour = "\033[35m"
	blue   terminalColour = "\033[34m"
	cyan   terminalColour = "\033[36m"

	bold = "\033[1m"

	fileCreatedToken = "+"
	fileIgnoredToken = "|"
	eventToken       = "➤"
	connectorToken   = "   |\n"
)

// Package prints the name of the package.
func Package(name string) {
	fmt.Printf("📦 %s\n", atom(bold+purple, name+"/"))
}

// File prints the name of the file and its status.
func File(name string, created bool) {
	statusToken, statusColour := fileIgnoredToken, grey
	fileColour := statusColour
	if created {
		statusToken, statusColour = fileCreatedToken, green
		fileColour = white
	}
	fmt.Printf("%s  %s\n", atom(statusColour, statusToken), atom(fileColour, name))
}

// Report prints the final report.
func Report(began time.Time, total, pkgCount int) {
	if total == 0 {
		fmt.Println(eventPrefix("💭"), core.DomainDir, "unchanged")
	} else {
		fmt.Printf("%s Generated %s across %s in %s.\n",
			eventPrefix("🤓"),
			atom(blue, fmt.Sprintf("%d files", total)),
			atom(blue, fmt.Sprintf("%d packages", pkgCount)),
			atom(cyan, time.Since(began).String()),
		)
	}
}

var connectorAtom = atom(grey, connectorToken)

// Info prints a message.
func Info(lines ...string) {
	fmt.Println(infoAtom("💡", lines...))
}

var rxp = regexp.MustCompile("(Key:.+$)")

// Error prints an error message.
func Error(err error) {
	msg := err.Error()
	if valErrs, ok := err.(validator.ValidationErrors); ok {
		for _, valErr := range valErrs {

			if _, ok := valErr.(validator.FieldError); ok {
				// TODO
			}
		}
	}
	fmt.Printf("%s%s\n",
		atom(red, eventPrefix("🫣 "), "You've encountered an error:", msg),
		infoAtom("🐞",
			fmt.Sprintf("Please check the error log file %s for more details.", atom(bold, logFile)),
			fmt.Sprintf("If the issue persists, please do report it to me: %s 👈", atom(cyan, internal.GHIssuesURL)),
		))
}

func infoAtom(emoji string, lines ...string) string {
	return fmt.Sprintf("\n%s%s %s",
		connectorAtom,
		eventPrefix(emoji),
		strings.Join(lines, "\n     "))
}

func atom(colour terminalColour, text ...string) string {
	return fmt.Sprintf("%s%s%s", colour, strings.Join(text, " "), white)
}

func eventPrefix(emoji string) string {
	return emoji + " " + eventToken
}
