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
package customer

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateRandomName creates a random full name using the firstNames and lastNames slices.
// This helper function extracts the name generation logic for testing.
func generateRandomName() string {
	randSource := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randSource)
	firstName := firstNames[randGen.Intn(len(firstNames))]
	lastName := lastNames[randGen.Intn(len(lastNames))]
	return fmt.Sprintf("%s %s", firstName, lastName)
}

func TestFirstNamesNotEmpty(t *testing.T) {
	require.NotEmpty(t, firstNames, "firstNames slice should not be empty")
}

func TestLastNamesNotEmpty(t *testing.T) {
	require.NotEmpty(t, lastNames, "lastNames slice should not be empty")
}

func TestGenerateRandomNameFormat(t *testing.T) {
	name := generateRandomName()

	// Name should contain a space (first + last)
	assert.Contains(t, name, " ", "generated name should contain a space between first and last name")

	// Name should have exactly two parts
	parts := strings.Split(name, " ")
	assert.Len(t, parts, 2, "generated name should have exactly two parts")
}

func TestGenerateRandomNameUsesValidNames(t *testing.T) {
	name := generateRandomName()
	parts := strings.Split(name, " ")

	// Verify first name is from the firstNames slice
	firstName := parts[0]
	assert.Contains(t, firstNames, firstName, "first name should be from firstNames slice")

	// Verify last name is from the lastNames slice
	lastName := parts[1]
	assert.Contains(t, lastNames, lastName, "last name should be from lastNames slice")
}

func TestGenerateRandomNameProducesVariety(t *testing.T) {
	// Generate multiple names and verify we get at least some variety
	names := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name := generateRandomName()
		names[name] = true
	}

	// With 16 first names and 16 last names (256 combinations),
	// 100 samples should produce at least a few unique names
	assert.Greater(t, len(names), 1, "should generate varied names")
}
