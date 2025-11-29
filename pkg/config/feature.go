package config

type FeatureFlag struct {
	Authentication string // disable, postgres
	Product        string // disable, postgres, mongo
}

func (f *FeatureFlag) Load() {
	f = &FeatureFlag{
		Authentication: getenv("FEATURE_AUTHENTICATION", "disable"),
		Product:        getenv("FEATURE_PRODUCT", "disable"),
	}
}
