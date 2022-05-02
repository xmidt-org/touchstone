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

package touchstone

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.uber.org/fx"
)

func Example() {
	type In struct {
		fx.In
		Counter  prometheus.Counter   `name:"example_counter"`
		GaugeVec *prometheus.GaugeVec `name:"example_gaugevec"`
	}

	_ = fx.New(
		fx.Logger(log.New(ioutil.Discard, "", 0)),
		fx.Supply(
			// this Config instance can come from anywhere
			Config{
				DefaultNamespace: "example",
				DefaultSubsystem: "touchstone",
			},
		),
		Provide(),
		Counter(prometheus.CounterOpts{
			// this will use the default namespace and subsystem
			Name: "example_counter",
		}),
		GaugeVec(prometheus.GaugeOpts{
			Subsystem: "override",
			Name:      "example_gaugevec",
		}, "label"),
		fx.Invoke(
			func(in In) {
				in.Counter.Inc()
				in.GaugeVec.WithLabelValues("value").Add(2.0)

				fmt.Println("counter", testutil.ToFloat64(in.Counter))
				fmt.Println("gaugevec", testutil.ToFloat64(in.GaugeVec))
			},
		),
	)

	// Output:
	// counter 1
	// gaugevec 2
}
