package touchtest

import "testing"

// mockTestingT is a mock/stub require.TestingT used to verify
// the correct errors happen.
type mockTestingT struct {
	t *testing.T

	errors   int
	failures int
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errors++
	m.t.Logf("<DEBUG> "+format+" <DEBUG>", args...)
}

func (m *mockTestingT) FailNow() {
	m.failures++
	m.t.Logf("<DEBUG> FailNow <DEBUG>")
}
