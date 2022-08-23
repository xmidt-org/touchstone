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
