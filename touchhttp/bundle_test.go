// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchhttp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/touchstone"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type BundleSuite struct {
	suite.Suite

	// now is a known start time for all clocks
	now time.Time
}

func (suite *BundleSuite) SetupTest() {
	suite.now = time.Now()
}

type ServerBundleSuite struct {
	BundleSuite
}

func (suite *ServerBundleSuite) testNewInstrumenterDefaults() {
	var (
		si ServerInstrumenter

		app = fxtest.New(
			suite.T(),
			touchstone.Provide(),
			fx.Provide(
				ServerBundle{}.NewInstrumenter(),
			),
			fx.Populate(&si),
		)
	)

	app.RequireStart()
	app.RequireStop()
}

func (suite *ServerBundleSuite) testNewInstrumenterNamed() {
	var (
		sb ServerBundle

		app = fxtest.New(
			suite.T(),
			touchstone.Provide(),
			fx.Provide(
				fx.Annotated{
					Name: "servers.main",
					Target: sb.NewInstrumenter(
						ServerLabel, "servers.main",
					),
				},
				fx.Annotated{
					Name: "servers.health",
					Target: sb.NewInstrumenter(
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

func (suite *ServerBundleSuite) TestNewInstrumenter() {
	suite.Run("Defaults", suite.testNewInstrumenterDefaults)
	suite.Run("Named", suite.testNewInstrumenterNamed)
}

func TestServerBundle(t *testing.T) {
	suite.Run(t, new(ServerBundleSuite))
}

type ClientBundleSuite struct {
	BundleSuite
}

func (suite *ClientBundleSuite) testNewInstrumenterDefaults() {
	var (
		ci ClientInstrumenter

		app = fxtest.New(
			suite.T(),
			touchstone.Provide(),
			fx.Provide(
				ClientBundle{}.NewInstrumenter(),
			),
			fx.Populate(&ci),
		)
	)

	app.RequireStart()
	app.RequireStop()
}

func (suite *ClientBundleSuite) testNewInstrumenterNamed() {
	var (
		cb ClientBundle

		app = fxtest.New(
			suite.T(),
			touchstone.Provide(),
			fx.Provide(
				fx.Annotated{
					Name: "clients.main",
					Target: cb.NewInstrumenter(
						ClientLabel, "clients.main",
					),
				},
				fx.Annotated{
					Name: "clients.consul",
					Target: cb.NewInstrumenter(
						ClientLabel, "clients.consul",
					),
				},
			),
			fx.Invoke(
				fx.Annotate(
					func(ClientInstrumenter) {},
					fx.ParamTags(`name:"clients.main"`),
				),
				fx.Annotate(
					func(ClientInstrumenter) {},
					fx.ParamTags(`name:"clients.consul"`),
				),
			),
		)
	)

	app.RequireStart()
	app.RequireStop()
}

func (suite *ClientBundleSuite) TestNewInstrumenter() {
	suite.Run("Defaults", suite.testNewInstrumenterDefaults)
	suite.Run("Named", suite.testNewInstrumenterNamed)
}

func TestClientBundle(t *testing.T) {
	suite.Run(t, new(ClientBundleSuite))
}
