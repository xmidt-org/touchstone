// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchstone

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Config defines the configuration options for bootstrapping a prometheus-based metrics environment.
type Config struct {
	// DefaultNamespace is the prometheus namespace to apply when a metric has no namespace.
	DefaultNamespace string `json:"defaultNamespace" yaml:"defaultNamespace"`

	// DefaultSubsystem is the prometheus subsystem to apply when a metric has no subsystem.
	DefaultSubsystem string `json:"defaultSubsystem" yaml:"defaultSubsystem"`

	// Pedantic controls whether a pedantic Registerer is used as the prometheus backend.
	//
	// See: https://godoc.org/github.com/prometheus/client_golang/prometheus#NewPedanticRegistry
	Pedantic bool `json:"pedantic" yaml:"pedantic"`

	// DisableGoCollector controls whether the go collector is registered on startup.
	// By default, the go collector is registered.
	//
	// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/collectors#NewGoCollector
	DisableGoCollector bool `json:"disableGoCollector" yaml:"disableGoCollector"`

	// DisableProcessCollector controls whether the process collector is registered on startup.
	// By default, this collector is registered.
	//
	// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/collectors#NewProcessCollector
	DisableProcessCollector bool `json:"disableProcessCollector" yaml:"disableProcessCollector"`

	// DisableBuildInfoCollector controls whether the build info collector is registered on startup.
	// By default, this collector is registered.
	//
	// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/collectors#NewBuildInfoCollector
	DisableBuildInfoCollector bool `json:"disableBuildInfoCollector" yaml:"disableBuildInfoCollector"`
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
		err = pr.Register(collectors.NewGoCollector())
	}

	if err == nil && !cfg.DisableProcessCollector {
		err = pr.Register(
			collectors.NewProcessCollector(
				collectors.ProcessCollectorOpts{
					Namespace: cfg.DefaultNamespace,
				},
			),
		)
	}

	if err == nil && !cfg.DisableBuildInfoCollector {
		err = pr.Register(collectors.NewBuildInfoCollector())
	}

	if err == nil {
		g = pr
		r = pr
	}

	return
}
