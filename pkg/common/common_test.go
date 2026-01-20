/*
Copyright © 2024 Mattias Gees mattias.gees@venafi.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package common

import (
	"testing"
	"time"
)

func TestTimeFormatIsValid(t *testing.T) {
	// Verify the time format string is valid by parsing a formatted time
	now := time.Now()
	formatted := now.Format(TimeFormat)

	_, err := time.Parse(TimeFormat, formatted)
	if err != nil {
		t.Errorf("TimeFormat is invalid: %v", err)
	}
}

func TestTimeFormatProducesExpectedOutput(t *testing.T) {
	// Test with a known time
	testTime := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)
	formatted := testTime.Format(TimeFormat)

	expected := "15/03/24 14:30:45"
	if formatted != expected {
		t.Errorf("TimeFormat produced %q, expected %q", formatted, expected)
	}
}

func TestDefaultTimeoutIsPositive(t *testing.T) {
	if DefaultTimeout <= 0 {
		t.Errorf("DefaultTimeout should be positive, got %v", DefaultTimeout)
	}
}

func TestDefaultTimeoutValue(t *testing.T) {
	expected := 3 * time.Second
	if DefaultTimeout != expected {
		t.Errorf("DefaultTimeout is %v, expected %v", DefaultTimeout, expected)
	}
}
