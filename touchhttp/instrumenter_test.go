// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package touchhttp

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/touchstone"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type NewServerInstrumenterSuite struct {
	suite.Suite
}

func (suite *NewServerInstrumenterSuite) TestDefaults() {
	var (
		si ServerInstrumenter

		app = fxtest.New(
			suite.T(),
			touchstone.Provide(),
			fx.Provide(
				NewServerInstrumenter(),
			),
			fx.Populate(&si),
		)
	)

	app.RequireStart()
	app.RequireStop()
}

func (suite *NewServerInstrumenterSuite) TestNamed() {
	var (
		app = fxtest.New(
			suite.T(),
			touchstone.Provide(),
			fx.Provide(
				fx.Annotated{
					Name: "servers.main",
					Target: NewServerInstrumenter(
						ServerLabel, "servers.main",
					),
				},
				fx.Annotated{
					Name: "servers.health",
					Target: NewServerInstrumenter(
						ServerLabel, "servers.health",
					),
				},
			),
			fx.Invoke(
				fx.Annotate(
					func(ServerInstrumenter) {},
					fx.ParamTags(`name:"servers.main"`),
				),
				fx.Annotate(
					func(ServerInstrumenter) {},
					fx.ParamTags(`name:"servers.health"`),
				),
			),
		)
	)

	app.RequireStart()
	app.RequireStop()
}

func TestNewServerInstrumenter(t *testing.T) {
	suite.Run(t, new(NewServerInstrumenterSuite))
}
