package options

// Options are configuration options that can be set by Environment Variables
type Options struct {
	// General
	Version string `envconfig:"VERSION" required:"true"`

	// Kubernetes
	// IsInCluster - Whether to use in cluster communication (if deployed inside of Kubernetes) or to look for a kubeconfig in home directory
	IsInCluster bool `envconfig:"IS_IN_CLUSTER" default:"true"`

	// Logger
	// LogLevel - Logger's log granularity (debug, info, warn, error, fatal, panic)
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

// NewOptions provides Application Options
func NewOptions() *Options {
	return &Options{}
}
