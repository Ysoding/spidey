package config

type SpideyConfig struct {
	TargetURL           string
	EnableCheckExternal bool
}

func NewDefaultConfig() SpideyConfig {
	return SpideyConfig{
		TargetURL: DEFALT_TARGET_URL,
	}
}
