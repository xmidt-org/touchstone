// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchhttp_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/xmidt-org/httpaux/client"
	"github.com/xmidt-org/touchstone"
	"github.com/xmidt-org/touchstone/touchhttp"
	"go.uber.org/fx"
)

// ExampleProvide shows how to bootstrap the http.Handler that serves
// prometheus scrape output.  touchhttp.Provide supplies a touchhttp.Handler
// that can be mounted on any router.
func ExampleProvide() {
	var h touchhttp.Handler
	app := fx.New(
		fx.NopLogger,
		touchstone.Provide(),
		touchhttp.Provide(),
		fx.Populate(&h),
	)

	if err := app.Err(); err != nil {
		fmt.Println(err)
		return
	}

	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	fmt.Println(response.Code)
	fmt.Println(strings.Contains(response.Body.String(), "go_goroutines"))

	// Output:
	// 200
	// true
}

// ExampleServerBundle shows how to instrument an http.Handler.  The
// ServerInstrumenter's Then method is a standard middleware, compatible
// with justinas/alice and gorilla/mux.  The server label distinguishes
// metrics from multiple servers in the same application.
func ExampleServerBundle() {
	var (
		g  prometheus.Gatherer
		si touchhttp.ServerInstrumenter
	)

	app := fx.New(
		fx.NopLogger,
		touchstone.Provide(),
		fx.Provide(
			touchhttp.ServerBundle{}.NewInstrumenter(touchhttp.ServerLabel, "main"),
		),
		fx.Populate(&g, &si),
	)

	if err := app.Err(); err != nil {
		fmt.Println(err)
		return
	}

	handler := si.Then(
		http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.WriteHeader(http.StatusAccepted)
		}),
	)

	for range 2 {
		handler.ServeHTTP(
			httptest.NewRecorder(),
			httptest.NewRequest(http.MethodPost, "/", nil),
		)
	}

	expected := `
# HELP server_request_count the total number of requests received since startup
# TYPE server_request_count counter
server_request_count{code="202",method="POST",server="main"} 2
`

	fmt.Println(testutil.GatherAndCompare(g, strings.NewReader(expected), touchhttp.DefaultServerCount))

	// Output:
	// <nil>
}

// ExampleClientBundle shows how to instrument an HTTP client.  The
// ClientInstrumenter's Then method decorates any httpaux.Client, which
// *http.Client implements.  Transactions that return an error are counted
// separately, and a missing response is recorded with a code of -1.
func ExampleClientBundle() {
	var (
		g  prometheus.Gatherer
		ci touchhttp.ClientInstrumenter
	)

	app := fx.New(
		fx.NopLogger,
		touchstone.Provide(),
		fx.Provide(
			touchhttp.ClientBundle{}.NewInstrumenter(touchhttp.ClientLabel, "api"),
		),
		fx.Populate(&g, &ci),
	)

	if err := app.Err(); err != nil {
		fmt.Println(err)
		return
	}

	// a stand-in for a real *http.Client that fails every other request
	calls := 0
	fake := client.Func(func(*http.Request) (*http.Response, error) {
		calls++
		if calls%2 == 0 {
			return nil, errors.New("connection refused")
		}

		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	})

	instrumented := ci.Then(fake)
	for range 4 {
		request := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		if response, err := instrumented.Do(request); err == nil {
			response.Body.Close()
		}
	}

	expected := `
# HELP client_request_count the total number of requests sent since startup
# TYPE client_request_count counter
client_request_count{client="api",code="-1",method="GET"} 2
client_request_count{client="api",code="200",method="GET"} 2
# HELP client_error_count the total number of errors (nil responses) since startup
# TYPE client_error_count counter
client_error_count{client="api",code="-1",method="GET"} 2
`

	fmt.Println(testutil.GatherAndCompare(
		g,
		strings.NewReader(expected),
		touchhttp.DefaultClientCount,
		touchhttp.DefaultClientErrorCount,
	))

	// Output:
	// <nil>
}
