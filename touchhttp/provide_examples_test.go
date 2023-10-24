// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchhttp

func Example() {
	/*
		_ = fx.New(
			fx.Logger(log.New(ioutil.Discard, "", 0)),
			touchstone.Provide(),
			Provide(),
			fx.Provide(
				func(sb ServerBundle) func(http.Handler) http.Handler { // can use justinas/alice
					return sb.ForServer("main").Then
				},
				func(middleware func(http.Handler) http.Handler) http.Handler {
					return middleware(
						http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
							fmt.Println("handled!")
						}),
					)
				},
			),
			fx.Invoke(
				func(h http.Handler) {
					h.ServeHTTP(
						httptest.NewRecorder(),
						httptest.NewRequest("GET", "/", nil),
					)
				},
			),
		)

		// Output:
		// handled!
	*/
}
