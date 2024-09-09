//go:build unit

// Package buildinfo provides functionality to store and retrieve information about a build,
// such as version, date, and commit hash. It includes a struct to hold this information and
// functions to create and format instances of this struct.
package buildinfo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testVersion = "0.0.1"
	testDate    = "2024-01-01"
	testCommit  = "048e08d569ecbfb1d7d6ee6acfd4ab052d923f39"
	na          = "N/A"
)

func Test(t *testing.T) {
	t.Run("not empty test", func(t *testing.T) {
		want := fmt.Sprintf("Build version: %s; Build date: %s; Build commit: %s;",
			testVersion, testDate, testCommit)
		bi := New(testVersion, testDate, testCommit)
		assert.Equal(t, want, bi.String())
	})

	t.Run("empty test", func(t *testing.T) {
		want := fmt.Sprintf("Build version: %s; Build date: %s; Build commit: %s;",
			na, na, na)
		bi := New("", "", "")
		assert.Equal(t, want, bi.String())
	})
}
