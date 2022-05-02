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
