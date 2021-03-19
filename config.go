package touchstone

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Config defines the configuration options for bootstrapping a prometheus-based metrics environment.
type Config struct {
	// DefaultNamespace is the prometheus namespace to apply when a metric has no namespace.
	DefaultNamespace string `json:"defaultNamespace" yaml:"defaultNamespace"`

	// DefaultSubsystem is the prometheus subsystem to apply when a metric has no subsystem.
	DefaultSubsystem string `json:"defaultSubsystem" yaml:"defaultSubsystem"`

	// Pedantic controls whether a pedantic Registerer is used as the prometheus backend.
	//
	// See https://godoc.org/github.com/prometheus/client_golang/prometheus#NewPedanticRegistry
	Pedantic bool `json:"pedantic" yaml:"pedantic"`

	// DisableGoCollector controls whether the go collector is registered on startup.
	// By default, the go collector is registered.
	//
	// See https://godoc.org/github.com/prometheus/client_golang/prometheus#NewGoCollector
	DisableGoCollector bool `json:"disableGoCollector" yaml:"disableGoCollector"`

	// DisableProcessCollector controls whether the process collector is registered on startup.
	// By default, this collector is registered.
	//
	// See https://godoc.org/github.com/prometheus/client_golang/prometheus#NewProcessCollector
	DisableProcessCollector bool `json:"disableProcessCollector" yaml:"disableProcessCollector"`
}

// New bootstraps a prometheus registry given a Config instance.  Note that the
// returned Registerer may be decorated to arbitrary depth.
func New(cfg Config) (g prometheus.Gatherer, r prometheus.Registerer, err error) {
	var pr *prometheus.Registry
	if cfg.Pedantic {
		pr = prometheus.NewPedanticRegistry()
	} else {
		pr = prometheus.NewRegistry()
	}

	if !cfg.DisableGoCollector {
		err = pr.Register(prometheus.NewGoCollector())
	}

	if err == nil && !cfg.DisableProcessCollector {
		err = pr.Register(
			prometheus.NewProcessCollector(
				prometheus.ProcessCollectorOpts{
					Namespace: cfg.DefaultNamespace,
				},
			),
		)
	}

	if err == nil {
		g = pr
		r = pr
	}

	return
}
