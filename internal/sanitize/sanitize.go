package sanitize

func Secret(value string) string {
	if value == "" {
		return ""
	}
	return "***"
}
