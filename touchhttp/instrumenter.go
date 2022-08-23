/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package touchhttp

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/httpaux"
	"github.com/xmidt-org/httpaux/client"
	"github.com/xmidt-org/httpaux/observe"
	"github.com/xmidt-org/touchstone"
	"go.uber.org/fx"
)

// transaction represents a completed HTTP transaction.
type transaction struct {
	start       time.Time
	code        int
	method      string
	err         error // that came from a client
	requestSize int64
}

// instrumenter is the common logic that decorates HTTP transactions for
// both clients and servers.
type instrumenter struct {
	count       *prometheus.CounterVec
	inFlight    prometheus.Gauge
	requestSize prometheus.ObserverVec
	duration    prometheus.ObserverVec

	// only used in clients
	errorCount *prometheus.CounterVec

	now func() time.Time
}

// begin records the start of an HTTP transaction
func (i instrumenter) begin(r *http.Request) transaction {
	i.inFlight.Inc()
	return transaction{
		start:       i.now(),
		method:      r.Method,
		requestSize: r.ContentLength,
	}
}

func (i instrumenter) endHandle(sc observe.StatusCoder, t transaction) {
	t.code = sc.StatusCode()
	i.end(t)
}

func (i instrumenter) endDo(response *http.Response, err error, t transaction) {
	if response != nil {
		t.code = response.StatusCode
	} else {
		t.code = -1
	}

	t.err = err
	i.end(t)
}

// end records the end of an HTTP transaction
func (i instrumenter) end(t transaction) {
	i.inFlight.Dec()

	l := prometheus.Labels(NewLabels(t.code, t.method))

	i.count.With(l).Inc()
	elapsed := i.now().Sub(t.start)
	i.duration.With(l).Observe(
		float64(elapsed / time.Millisecond),
	)

	i.requestSize.With(l).Observe(
		float64(t.requestSize),
	)

	if i.errorCount != nil && t.err != nil {
		i.errorCount.With(l).Inc()
	}
}

// ServerInstrumenter is a serverside middleware that provides http.Handler
// metrics.
type ServerInstrumenter struct {
	instrumenter
}

// Then is a server middleware that instruments the given handler.  This middleware
// is compatible with justinas/alice and gorilla/mux.
func (si ServerInstrumenter) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		w := observe.New(rw)
		t := si.begin(r)
		defer si.endHandle(w, t)

		next.ServeHTTP(w, r)
	})
}

// ClientInstrumenter is a clientside middleware that provides HTTP client
// metrics.
type ClientInstrumenter struct {
	instrumenter
}

// Then is a client middleware that instruments the given client.  This middleware
// is compatible with httpaux.
func (ci ClientInstrumenter) Then(next httpaux.Client) httpaux.Client {
	return client.Func(func(request *http.Request) (response *http.Response, err error) {
		t := ci.begin(request)
		response, err = next.Do(request)
		ci.endDo(response, err, t)
		return
	})
}

var _ client.Constructor = ClientInstrumenter{}.Then

// ServerInstrumenterIn defines the set of dependencies required to build a ServerInstrumenter.
type ServerInstrumenterIn struct {
	fx.In

	// Factory is the required touchstone Factory instance.
	Factory *touchstone.Factory

	// Bundle is the optional ServerBundle supplied in the application.
	// If not present, the default metrics are used.
	Bundle ServerBundle `optional:"true"`
}

// NewServerInstrumenter produces a constructor that can be passed to fx.Provide.  The returned
// constructor allows a ServerBundle to be injected.
//
// Use this function when a ServerBundle has been supplied to the enclosing fx.App:
//
//   app := fx.New(
//     touchstone.Provide(), // bootstrap metrics subsystem
//
//     fx.Provide(
//       // A single, global ServerInstrumenter
//       touchhttp.NewServerInstrumenter(),
//
//       // A custom label
//       touchhttp.NewServerInstrumenter(
//         "custom1", "value",
//       ),
//
//       // A named ServerInstrumenter with a server label
//       fx.Annotated{
//         Name: "servers.main",
//         Target: NewServerInstrumenter(
//           touchhttp.ServerLabel, "servers.main",
//         ),
//       },
//     ),
//   )
func NewServerInstrumenter(namesAndValues ...string) func(ServerInstrumenterIn) (ServerInstrumenter, error) {
	return func(in ServerInstrumenterIn) (ServerInstrumenter, error) {
		return in.Bundle.NewInstrumenter(
			namesAndValues...,
		)(in.Factory)
	}
}
