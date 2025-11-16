package telemetry

import (
	"errors"
	"time"
)

const (
	defaultServiceVersion  = "local"
	defaultEnvironment     = "development"
	defaultExporterTimeout = 10 * time.Second
	defaultExportInterval  = 60 * time.Second
)

// Options contains shared configuration for OpenTelemetry providers.
type Options struct {
	serviceName     string
	serviceVersion  string
	environment     string
	exporterURL     string
	insecure        bool
	exporterTimeout time.Duration
	exportInterval  time.Duration
}

// Option is a functional option for configuring OpenTelemetry providers.
type Option func(*Options) error

// WithServiceName sets the service name (required).
func WithServiceName(name string) Option {
	return func(o *Options) error {
		if name == "" {
			return errors.New("service name cannot be empty")
		}
		o.serviceName = name
		return nil
	}
}

// WithServiceVersion sets the service version.
// Defaults to "unknown" if not set.
func WithServiceVersion(version string) Option {
	return func(o *Options) error {
		if version == "" {
			version = defaultServiceVersion
		}
		o.serviceVersion = version
		return nil
	}
}

// WithEnvironment sets the deployment environment.
// Defaults to "development" if not set.
func WithEnvironment(env string) Option {
	return func(o *Options) error {
		if env == "" {
			env = defaultEnvironment
		}
		o.environment = env
		return nil
	}
}

// WithExporterURL sets the OTLP exporter endpoint URL (required).
func WithExporterURL(url string) Option {
	return func(o *Options) error {
		if url == "" {
			return errors.New("exporter URL cannot be empty")
		}
		o.exporterURL = url
		return nil
	}
}

// WithInsecure sets whether to use insecure connection.
// Should be true for local development, false for production.
func WithInsecure(insecure bool) Option {
	return func(o *Options) error {
		o.insecure = insecure
		return nil
	}
}

// WithExporterTimeout sets the timeout for exporter operations.
// Defaults to 10 seconds if not set or zero.
func WithExporterTimeout(timeout time.Duration) Option {
	return func(o *Options) error {
		if timeout == 0 {
			timeout = defaultExporterTimeout
		}
		o.exporterTimeout = timeout
		return nil
	}
}

// WithExportInterval sets the interval for periodic metric exports.
// Defaults to 60 seconds if not set or zero.
// Only used by metrics provider.
func WithExportInterval(interval time.Duration) Option {
	return func(o *Options) error {
		if interval == 0 {
			interval = defaultExportInterval
		}
		o.exportInterval = interval
		return nil
	}
}

// NewOptions creates a new Options with the provided functional options.
func NewOptions(opts ...Option) (Options, error) {
	o := Options{
		serviceVersion:  defaultServiceVersion,
		environment:     defaultEnvironment,
		exporterTimeout: defaultExporterTimeout,
		exportInterval:  defaultExportInterval,
	}

	for _, opt := range opts {
		if err := opt(&o); err != nil {
			return Options{}, err
		}
	}

	// Validate required fields
	if o.serviceName == "" {
		return Options{}, errors.New("service name is required")
	}
	if o.exporterURL == "" {
		return Options{}, errors.New("exporter URL is required")
	}

	return o, nil
}

// Getters for accessing private fields

func (o Options) ServiceName() string {
	return o.serviceName
}

func (o Options) ServiceVersion() string {
	return o.serviceVersion
}

func (o Options) Environment() string {
	return o.environment
}

func (o Options) ExporterURL() string {
	return o.exporterURL
}

func (o Options) Insecure() bool {
	return o.insecure
}

func (o Options) ExporterTimeout() time.Duration {
	return o.exporterTimeout
}

func (o Options) ExportInterval() time.Duration {
	return o.exportInterval
}
