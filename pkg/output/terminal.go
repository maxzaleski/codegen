package output

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

	fileCreatedToken      = "+"
	fileIgnoredToken      = "|"
	eventToken            = "➤"
	connectorTokenFile    = "   |\n"
	connectorToken        = "├─"
	connectorTokenFinal   = "└─"
	connectorTokenNeutral = "│"
)
