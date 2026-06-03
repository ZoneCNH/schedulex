package templatex

import (
	"errors"
	"time"

	"github.com/ZoneCNH/xlib-standard/internal/sanitize"
	"github.com/ZoneCNH/xlib-standard/internal/validation"
)

type Config struct {
	Name    string
	Timeout time.Duration
	Secret  string
}

type SanitizedConfig struct {
	Name    string
	Timeout time.Duration
	Secret  string
}

func (c Config) Validate() error {
	if err := validation.RequireNonEmpty("name", c.Name); err != nil {
		return validationError("Config.Validate", err.Error(), err)
	}
	if c.Timeout < 0 {
		err := errors.New("timeout must not be negative")
		return validationError("Config.Validate", err.Error(), err)
	}
	return nil
}

func (c Config) Sanitize() SanitizedConfig {
	return SanitizedConfig{
		Name:    c.Name,
		Timeout: c.Timeout,
		Secret:  sanitize.Secret(c.Secret),
	}
}
