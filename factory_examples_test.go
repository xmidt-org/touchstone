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

func ExampleFactory() {
	var f *Factory
	_ = fx.New(
		fx.Logger(log.New(ioutil.Discard, "", 0)),
		Provide(),
		fx.Populate(&f),
	)

	c, err := f.NewCounterVec(
		prometheus.CounterOpts{
			Name: "example",
			Help: "here is a lovely example",
		}, "label",
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	c.WithLabelValues("value").Inc()
	fmt.Println(testutil.ToFloat64(c))

	// Output:
	// 1
}
