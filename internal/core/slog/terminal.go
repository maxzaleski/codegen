package slog

import (
	"fmt"
	"strings"
)

// Colour represents a terminal colour.
type Colour string

const (
	white Colour = "\033[0m"

	Grey   Colour = "\033[90m"
	White  Colour = "\033[0m"
	Green  Colour = "\033[32m"
	Red    Colour = "\033[31m"
	Purple Colour = "\033[35m"
	Blue   Colour = "\033[34m"
	Cyan   Colour = "\033[36m"
	Yellow Colour = "\033[33m"
	Pink   Colour = "\033[95m"

	Bold = "\033[1m"
)

func Atom(colour Colour, text ...string) string {
	return fmt.Sprintf("%s%s%s", colour, strings.Join(text, " "), white)
}

// domain returns a string in the format of `[domain:value]`.
func domain(tc Colour, domain, value string) string {
	return fmt.Sprintf("[%s:%s]", domain, Atom(tc, value))
}
