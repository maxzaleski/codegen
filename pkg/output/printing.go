package output

import (
	"fmt"
	"time"
)

// termColour represents a terminal colour.
type termColour string

const (
	grey   termColour = "\033[90m"
	white  termColour = "\033[0m"
	green  termColour = "\033[32m"
	red    termColour = "\033[31m"
	purple termColour = "\033[35m"
	blue   termColour = "\033[34m"
	cyan   termColour = "\033[36m"

	bold = "\033[1m"

	fileCreatedToken = "+"
	fileIgnoredToken = "|"
	eventToken       = "➤"
)

// Package prints the name of the package.
func Package(name string) {
	fmt.Printf("📦 %s\n", atom(bold+purple, name+"/"))
}

// File prints the name of the file and its status.
func File(name string, created bool) {
	statusToken := fileIgnoredToken
	statusColour := grey
	fileColour := statusColour
	if created {
		statusToken = fileCreatedToken
		statusColour = green
		fileColour = white
	}

	fmt.Printf("%s  %s\n", atom(statusColour, statusToken), atom(fileColour, name))
}

// Report prints the final report.
func Report(began time.Time, total, pkgCount int) {
	pfx := "🤓 " + eventToken
	if total == 0 {
		fmt.Println(atom(blue, pfx+" No additional files were generated."))
	} else {
		fmt.Printf(pfx+" Generated %s across %s in %s.\n",
			atom(blue, fmt.Sprintf("%d files", total)),
			atom(blue, fmt.Sprintf("%d packages", pkgCount)),
			atom(cyan, time.Since(began).String()),
		)
	}
}

// Info prints a message.
func Info(msg string) {
	fmt.Println(atom(grey, "   |\n💡"), eventToken, msg)
}

func EventIndent(msg string, newline bool) string {
	msg = "     " + msg
	if newline {
		msg = "\n" + msg
	}
	return msg
}

// Error prints an error message.
func Error(err error) {
	fmt.Printf("%s\n\n%s\n",
		atom(purple, "😡 ➤ An error has occurred:"),
		atom(red, fmt.Sprintf("%+v", err)))
}

func atom(colour termColour, text string) string {
	return fmt.Sprintf("%s%s%s", colour, text, white)
}
