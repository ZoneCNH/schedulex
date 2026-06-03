package validation

import "fmt"

func RequireNonEmpty(field string, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}
