package touchkit

import (
	"github.com/go-kit/kit/metrics"
	kitprom "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/touchstone"
)

// Factory handles creation and registration of go-kit metrics using prometheus
// as the backend.  This type fulfills the same purpose as touchstone.Factory:
// it allows easy creation of metrics combined with registration against an
// injected prometheus.Registerer.
//
// See: https://pkg.go.dev/github.com/xmidt-org/touchstone#Factory
type Factory struct {
	delegate *touchstone.Factory
}

// NewFactory produces a go-kit factory that delegates to the given touchstone Factory.
func NewFactory(f *touchstone.Factory) *Factory {
	return &Factory{
		delegate: f,
	}
}

// NewCounter creates a go-kit Counter using prometheus as the backend.
//
// NOTE: go-kit does not support plain prometheus counters.  A CounterVec is
// always created by this method and wrapped by go-kit.
func (f *Factory) NewCounter(o prometheus.CounterOpts, labelNames ...string) (m metrics.Counter, err error) {
	var cv *prometheus.CounterVec
	cv, err = f.delegate.NewCounterVec(o, labelNames...)
	if err == nil {
		m = kitprom.NewCounter(cv)
	}

	return
}

// NewGauge creates a go-kit Gauge using prometheus as the backend.
//
// NOTE: go-kit does not support plain prometheus gauges.  A GaugeVec is
// always created by this method and wrapped by go-kit.
func (f *Factory) NewGauge(o prometheus.GaugeOpts, labelNames ...string) (m metrics.Gauge, err error) {
	var gv *prometheus.GaugeVec
	gv, err = f.delegate.NewGaugeVec(o, labelNames...)
	if err == nil {
		m = kitprom.NewGauge(gv)
	}

	return
}

// NewHistogram creates a go-kit Histogram using prometheus as the backend.
//
// NOTE: go-kit does not support plain prometheus histogram.  A HistogramVec is
// always created by this method and wrapped by go-kit.
func (f *Factory) NewHistogram(o prometheus.HistogramOpts, labelNames ...string) (m metrics.Histogram, err error) {
	var observer prometheus.ObserverVec
	observer, err = f.delegate.NewHistogramVec(o, labelNames...)
	if err == nil {
		// TODO: this is really ugly.  is there a better way that avoids casting?
		m = kitprom.NewHistogram(observer.(*prometheus.HistogramVec))
	}

	return
}

// NewSummary creates a go-kit Summary using prometheus as the backend.  The go-kit metrics
// package has no Summary type.  Instead, prometheus summary vectors implement metrics.Histogram.
//
// NOTE: go-kit does not support plain prometheus summary.  A SummaryVec is
// always created by this method and wrapped by go-kit.
func (f *Factory) NewSummary(o prometheus.SummaryOpts, labelNames ...string) (m metrics.Histogram, err error) {
	var observer prometheus.ObserverVec
	observer, err = f.delegate.NewSummaryVec(o, labelNames...)
	if err == nil {
		// TODO: this is really ugly.  is there a better way that avoids casting?
		m = kitprom.NewSummary(observer.(*prometheus.SummaryVec))
	}

	return
}
