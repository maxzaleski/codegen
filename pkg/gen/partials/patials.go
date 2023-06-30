package partials

var goPartials = map[string]interface{}{}

func GetByExtension(ext string) map[string]interface{} {
	switch ext {
	case "go":
		return goPartials
	default:
		return nil
	}
}
