package spidey

import "net/http"

type Config struct {
	URL                 string
	EnableCheckExternal bool
	Client              *http.Client
	Events              Events
}

func NewDefaultConfig(events Events) Config {
	return Config{
		Client: http.DefaultClient,
		URL:    DEFAULT_TARGET_URL,
		Events: events,
	}
}
