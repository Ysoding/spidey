package spidey

import (
	"net/http"
	"time"
)

type Config struct {
	URL                 string
	EnableCheckExternal bool
	Client              *http.Client
	Events              Events
	Depth               int
}

func NewDefaultConfig(events Events) Config {
	return Config{
		Client:              &http.Client{Timeout: time.Second * 5},
		URL:                 DEFAULT_TARGET_URL,
		Events:              events,
		EnableCheckExternal: false,
		Depth:               2,
	}
}
