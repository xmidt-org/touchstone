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
