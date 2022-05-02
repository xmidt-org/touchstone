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
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type writeSyncer struct {
	t *testing.T
}

func (ws writeSyncer) Write(b []byte) (int, error) {
	ws.t.Log(string(b))
	return len(b), nil
}

func (ws writeSyncer) Sync() error { return nil }

// FxTestSuite provides common behaviors around the setup of a fxtest app
// for touchstone testing.
type FxTestSuite struct {
	suite.Suite

	logger *zap.Logger
}

func (suite *FxTestSuite) BeforeTest(suiteName, testName string) {
	suite.logger = zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(
				zapcore.EncoderConfig{
					MessageKey:  "msg",
					LevelKey:    "level",
					EncodeLevel: zapcore.LowercaseLevelEncoder,
				},
			),
			writeSyncer{t: suite.T()},
			zapcore.ErrorLevel,
		),
		zap.Fields(
			zap.String("suite", suiteName),
			zap.String("test", testName),
		),
	)
}

func (suite *FxTestSuite) newTestApp(options ...fx.Option) *fxtest.App {
	app := fxtest.New(
		suite.T(),
		append(
			[]fx.Option{
				fx.Supply(suite.logger),
			},
			options...,
		)...,
	)

	return app
}

func (suite *FxTestSuite) newApp(options ...fx.Option) *fx.App {
	app := fx.New(
		append(
			[]fx.Option{
				fx.Supply(suite.logger),
			},
			options...,
		)...,
	)

	return app
}
