// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package touchbundle

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/touchstone"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

type BundleSuite struct {
	suite.Suite
}

func (suite *BundleSuite) newFactory() *touchstone.Factory {
	cfg := touchstone.Config{
		Pedantic:                  true,
		DisableGoCollector:        true,
		DisableProcessCollector:   true,
		DisableBuildInfoCollector: true,
	}

	_, r, err := touchstone.New(cfg)
	suite.Require().NoError(err)
	suite.Require().NotNil(r)

	f := touchstone.NewFactory(cfg, zap.L(), r)
	suite.Require().NotNil(f)
	return f
}

func (suite *BundleSuite) successfulPopulate(bundle Bundle) {
	suite.Require().NoError(
		Populate(suite.newFactory(), bundle),
	)
}

func (suite *BundleSuite) testPopulateNonPointer() {
	type bundle struct {
		DoesNotMatter int
	}

	suite.Error(
		Populate(suite.newFactory(), bundle{}),
	)
}

func (suite *BundleSuite) testPopulateNonStruct() {
	suite.Error(
		Populate(suite.newFactory(), 123),
	)
}

func (suite *BundleSuite) testPopulateCounters() {
	type bundle struct {
		Counter1 prometheus.Counter
		Counter2 prometheus.Counter     `name:"custom2"`
		Counter3 *prometheus.CounterVec `labelNames:"foo,bar"`
		Counter4 *prometheus.CounterVec `name:"custom4" labelNames:"foo,bar"`
		Ignore1  prometheus.Counter     `touchstone:"-"`
		Ignore2  *prometheus.CounterVec `touchstone:"-"`
	}

	var b bundle
	suite.successfulPopulate(&b)
	suite.NotNil(b.Counter1, "Counter1 is nil")
	suite.NotNil(b.Counter2, "Counter2 is nil")
	suite.NotNil(b.Counter3, "Counter3 is nil")
	suite.NotNil(b.Counter4, "Counter4 is nil")
	suite.Nil(b.Ignore1, "Ignore1 is not nil")
	suite.Nil(b.Ignore2, "Ignore2 is not nil")

	suite.Run("LabelNamesOnNonVector", func() {
		type bundle struct {
			C prometheus.Counter `labelNames:"should,cause,an,error"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("MissingLabelNames", func() {
		type bundle struct {
			C *prometheus.CounterVec
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("EmptyLabelNames", func() {
		type bundle struct {
			C *prometheus.CounterVec `labelNames:""`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("InvalidTag", func() {
		type bundle struct {
			C prometheus.Counter `buckets:"1.0,2.0"` // invalid for a counter
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})
}

func (suite *BundleSuite) testPopulateGauges() {
	type bundle struct {
		Gauge1  prometheus.Gauge
		Gauge2  prometheus.Gauge     `name:"custom2"`
		Gauge3  *prometheus.GaugeVec `labelNames:"foo,bar"`
		Gauge4  *prometheus.GaugeVec `name:"custom4" labelNames:"foo,bar"`
		Ignore1 prometheus.Gauge     `touchstone:"-"`
		Ignore2 *prometheus.GaugeVec `touchstone:"-"`
	}

	var b bundle
	suite.successfulPopulate(&b)
	suite.NotNil(b.Gauge1, "Gauge1 is nil")
	suite.NotNil(b.Gauge2, "Gauge2 is nil")
	suite.NotNil(b.Gauge3, "Gauge3 is nil")
	suite.NotNil(b.Gauge4, "Gauge4 is nil")
	suite.Nil(b.Ignore1, "Ignore1 is not nil")
	suite.Nil(b.Ignore2, "Ignore2 is not nil")

	suite.Run("LabelNamesOnNonVector", func() {
		type bundle struct {
			G prometheus.Gauge `labelNames:"should,cause,an,error"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("MissingLabelNames", func() {
		type bundle struct {
			G *prometheus.GaugeVec
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("EmptyLabelNames", func() {
		type bundle struct {
			G *prometheus.GaugeVec `labelNames:""`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("InvalidTag", func() {
		type bundle struct {
			G prometheus.Gauge `buckets:"1.0,2.0"` // invalid for a gauge
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})
}

func (suite *BundleSuite) testPopulateHistograms() {
	type bundle struct {
		Histogram1 prometheus.Histogram
		Histogram2 prometheus.Histogram     `name:"custom2"`
		Histogram3 *prometheus.HistogramVec `labelNames:"foo,bar" buckets:"0.1, 0.2, 0.5, 1.0"`
		Histogram4 *prometheus.HistogramVec `name:"custom4" labelNames:"foo,bar"`
		Ignore1    prometheus.Histogram     `touchstone:"-"`
		Ignore2    *prometheus.HistogramVec `touchstone:"-"`
	}

	var b bundle
	suite.successfulPopulate(&b)
	suite.NotNil(b.Histogram1, "Histogram1 is nil")
	suite.NotNil(b.Histogram2, "Histogram2 is nil")
	suite.NotNil(b.Histogram3, "Histogram3 is nil")
	suite.NotNil(b.Histogram4, "Histogram4 is nil")
	suite.Nil(b.Ignore1, "Ignore1 is not nil")
	suite.Nil(b.Ignore2, "Ignore2 is not nil")

	suite.Run("LabelNamesOnNonVector", func() {
		type bundle struct {
			H prometheus.Histogram `labelNames:"should,cause,an,error"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("MissingLabelNames", func() {
		type bundle struct {
			H *prometheus.HistogramVec
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("EmptyLabelNames", func() {
		type bundle struct {
			H *prometheus.HistogramVec `labelNames:""`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("InvalidBuckets", func() {
		type bundle struct {
			H prometheus.Histogram `buckets:"this is not valid"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("AmbiguousTag", func() {
		type bundle struct {
			H prometheus.Histogram `buckets:"0.2, 0.5, 1.0, 2.0" bufCap:"123"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})
}

func (suite *BundleSuite) testPopulateSummaries() {
	type bundle struct {
		Summary1 prometheus.Summary
		Summary2 prometheus.Summary     `name:"custom2" objectives:"0.2:1.0, 0.5:2.0, 1.0:3.0" ageBuckets:"123" maxAge:"100m" bufCap:"1024"`
		Summary3 *prometheus.SummaryVec `labelNames:"foo,bar"`
		Summary4 *prometheus.SummaryVec `name:"custom4" labelNames:"foo,bar"`
		Ignore1  prometheus.Summary     `touchstone:"-"`
		Ignore2  *prometheus.SummaryVec `touchstone:"-"`
	}

	var b bundle
	suite.successfulPopulate(&b)
	suite.NotNil(b.Summary1, "Summary1 is nil")
	suite.NotNil(b.Summary2, "Summary2 is nil")
	suite.NotNil(b.Summary3, "Summary3 is nil")
	suite.NotNil(b.Summary4, "Summary4 is nil")
	suite.Nil(b.Ignore1, "Ignore1 is not nil")
	suite.Nil(b.Ignore2, "Ignore2 is not nil")

	suite.Run("LabelNamesOnNonVector", func() {
		type bundle struct {
			S prometheus.Summary `labelNames:"should,cause,an,error"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("MissingLabelNames", func() {
		type bundle struct {
			S *prometheus.SummaryVec
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("EmptyLabelNames", func() {
		type bundle struct {
			S *prometheus.SummaryVec `labelNames:""`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("InvalidObjectives", func() {
		type bundle struct {
			S prometheus.Summary `objectives:"this is not valid"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("InvalidMagAge", func() {
		type bundle struct {
			S prometheus.Summary `maxAge:"this is not valid"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("InvalidAgeBuckets", func() {
		type bundle struct {
			S prometheus.Summary `ageBuckets:"this is not valid"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("InvalidBufCap", func() {
		type bundle struct {
			S prometheus.Summary `bufCap:"this is not valid"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})
}

func (suite *BundleSuite) testPopulateObservers() {
	type bundle struct {
		Observer1 prometheus.Observer
		Observer2 prometheus.Observer `name:"custom2" buckets:"1.0,2.0,3.0"`
		Observer3 prometheus.Observer `objectives:"1.0:2.0"`
		Observer4 prometheus.Observer `type:"histogram"`
		Observer5 prometheus.Observer `type:"summary"`
		Ignore    prometheus.Observer `touchstone:"-"`
	}

	var b bundle
	suite.successfulPopulate(&b)
	suite.Implements((*prometheus.Histogram)(nil), b.Observer1)
	suite.Implements((*prometheus.Histogram)(nil), b.Observer2)
	suite.Implements((*prometheus.Summary)(nil), b.Observer3)
	suite.Implements((*prometheus.Histogram)(nil), b.Observer4)
	suite.Implements((*prometheus.Summary)(nil), b.Observer5)
	suite.Nil(b.Ignore)

	suite.Run("Ambiguous", func() {
		type bundle struct {
			O prometheus.Observer `buckets:"1.0, 2.0" bufCap:"123"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("LabelNames", func() {
		type bundle struct {
			O prometheus.Observer `labelNames:"not,allowed,for,nonvectors"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("InvalidType", func() {
		type bundle struct {
			O prometheus.Observer `type:"what is this?"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})
}

func (suite *BundleSuite) testPopulateObserverVecs() {
	type bundle struct {
		ObserverVec1 prometheus.ObserverVec `labelNames:"one"`
		ObserverVec2 prometheus.ObserverVec `name:"custom2" buckets:"1.0,2.0,3.0" labelNames:"foo,bar"`
		ObserverVec3 prometheus.ObserverVec `objectives:"1.0:2.0" labelNames:"foo"`
		ObserverVec4 prometheus.ObserverVec `type:"histogram" labelNames:"foo,bar,moo"`
		ObserverVec5 prometheus.ObserverVec `type:"summary" labelNames:"moo,mar,mab"`
		Ignore       prometheus.ObserverVec `touchstone:"-"`
	}

	var b bundle
	suite.successfulPopulate(&b)
	suite.IsType((*prometheus.HistogramVec)(nil), b.ObserverVec1)
	suite.IsType((*prometheus.HistogramVec)(nil), b.ObserverVec2)
	suite.IsType((*prometheus.SummaryVec)(nil), b.ObserverVec3)
	suite.IsType((*prometheus.HistogramVec)(nil), b.ObserverVec4)
	suite.IsType((*prometheus.SummaryVec)(nil), b.ObserverVec5)
	suite.Nil(b.Ignore)

	suite.Run("Ambiguous", func() {
		type bundle struct {
			O prometheus.ObserverVec `buckets:"1.0, 2.0" bufCap:"123"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("MissingLabelNames", func() {
		type bundle struct {
			O prometheus.ObserverVec
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})

	suite.Run("InvalidType", func() {
		type bundle struct {
			O prometheus.ObserverVec `type:"what is this?"`
		}

		var b bundle
		suite.Error(
			Populate(suite.newFactory(), &b),
		)
	})
}

func (suite *BundleSuite) TestPopulate() {
	suite.Run("NonPointer", suite.testPopulateNonPointer)
	suite.Run("NonStruct", suite.testPopulateNonStruct)
	suite.Run("Counters", suite.testPopulateCounters)
	suite.Run("Gauges", suite.testPopulateGauges)
	suite.Run("Histograms", suite.testPopulateHistograms)
	suite.Run("Summaries", suite.testPopulateSummaries)
	suite.Run("Observers", suite.testPopulateObservers)
	suite.Run("ObserverVecs", suite.testPopulateObserverVecs)
}

func (suite *BundleSuite) newApp(options ...fx.Option) *fx.App {
	app := fx.New(
		append(
			[]fx.Option{
				fx.WithLogger(func() fxevent.Logger {
					return fxtest.NewTestLogger(suite.T())
				}),
				fx.Supply(suite.newFactory()),
			},
			options...,
		)...,
	)

	suite.Require().NotNil(app)
	return app
}

func (suite *BundleSuite) newTestApp(options ...fx.Option) *fxtest.App {
	app := fxtest.New(
		suite.T(),
		append(
			[]fx.Option{
				fx.Supply(suite.newFactory()),
			},
			options...,
		)...,
	)

	suite.Require().NotNil(app)
	return app
}

func (suite *BundleSuite) testProvideInvalidPrototype() {
	app := suite.newApp(
		Provide(123),
	)

	suite.Error(app.Err())
}

func (suite *BundleSuite) testProvideStruct() {
	type bundle struct {
		C *prometheus.CounterVec `labelNames:"foo,bar"`
		G prometheus.Gauge       `name:"gauge"`
	}

	var b bundle
	app := suite.newTestApp(
		Provide(bundle{}),
		fx.Populate(&b),
	)

	app.RequireStart()
	app.RequireStop()
	suite.NotNil(b.C)
	suite.NotNil(b.G)

	suite.Run("InvalidBundle", func() {
		type bundle struct {
			C *prometheus.CounterVec // missing label names
		}

		var b bundle
		app := suite.newApp(
			Provide(bundle{}),
			fx.Populate(&b),
		)

		suite.Error(app.Err())
		suite.Nil(b.C)
	})
}

func (suite *BundleSuite) testProvidePointer() {
	type bundle struct {
		C *prometheus.CounterVec `labelNames:"foo,bar"`
		G prometheus.Gauge       `name:"gauge"`
	}

	var b *bundle
	app := suite.newTestApp(
		Provide(&bundle{}),
		fx.Populate(&b),
	)

	app.RequireStart()
	app.RequireStop()
	suite.Require().NotNil(b)
	suite.NotNil(b.C)
	suite.NotNil(b.G)

	suite.Run("InvalidBundle", func() {
		type bundle struct {
			C *prometheus.CounterVec // missing label names
		}

		var b *bundle
		app := suite.newApp(
			Provide(&bundle{}),
			fx.Populate(&b),
		)

		suite.Error(app.Err())
		suite.Nil(b)
	})
}

func (suite *BundleSuite) TestProvide() {
	suite.Run("InvalidPrototype", suite.testProvideInvalidPrototype)
	suite.Run("Struct", suite.testProvideStruct)
	suite.Run("Pointer", suite.testProvidePointer)
}

func TestBundle(t *testing.T) {
	suite.Run(t, new(BundleSuite))
}
