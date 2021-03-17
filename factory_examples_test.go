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
