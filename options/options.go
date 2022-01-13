package options

// Options are configuration options that can be set by Environment Variables
type Options struct {
	// General
	Version string `envconfig:"VERSION" default:"1.0.0"`

	// Kubernetes
	// IsInCluster - Whether to use in cluster communication (if deployed inside of Kubernetes) or to look for a kubeconfig in home directory
	IsInCluster bool `envconfig:"IS_IN_CLUSTER" default:"false"`

	// Logger
	// LogLevel - Logger's log granularity (debug, info, warn, error, fatal, panic)
	LogLevel string `envconfig:"LOG_LEVEL" default:"debug"`
}

// NewOptions provides Application Options
func NewOptions() *Options {
	return &Options{}
}
