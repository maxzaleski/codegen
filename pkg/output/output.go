package output

import "fmt"

const (
	grey   = "\033[90m"
	white  = "\033[0m"
	green  = "\033[32m"
	red    = "\033[31m"
	purple = "\033[35m"
	blue   = "\033[34m"

	bold = "\033[1m"

	fileCreatedToken = "+"
	fileIgnoredToken = "|"
)

// Package prints the name of the package.
func Package(name string) {
	fmt.Printf("📦 %s\n", atom(bold+purple, name+"/"))
}

// File prints the name of the file and its status.
func File(key string, created bool) {
	statusToken := fileIgnoredToken
	statusColour := grey
	fileColour := statusColour
	if created {
		statusToken = fileCreatedToken
		statusColour = green
		fileColour = white
	}

	fmt.Printf("%s  %s\n", atom(statusColour, statusToken), atom(fileColour, key))
}

// Report prints the final report.
func Report(total, pkgCount int) {
	if total == 0 {
		fmt.Println("🤓 No additional files were generated.")
	} else {
		fmt.Printf("🤓 Generated %s across %s.\n",
			atom(blue, fmt.Sprintf("%d files", total)),
			atom(blue, fmt.Sprintf("%d packages", pkgCount)),
		)
	}
}

// Info prints a message.
func Info(msg string) {
	fmt.Println(atom(blue, "   ↪ "+msg))
}

// Error prints an error message.
func Error(err error) {
	fmt.Println(atom(red, fmt.Sprintf("An error has occurred:\n\n%+v\n", err)))
}

func atom(colour, text string) string {
	return fmt.Sprintf("%s%s%s", colour, text, white)
}
