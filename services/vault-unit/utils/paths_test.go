package utils

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionString(t *testing.T) {
	testPad := func(version int) string {
		return fmt.Sprintf("%010d", version)
	}

	versionMin := testPad(0)
	versionMax := testPad(math.MaxInt32)

	assert.Equal(t, len(versionMin), len(versionMax))
}

func TestEventPath(t *testing.T) {
	account := "account_2"

	versionMin := 0
	pathMin := EventPath(account, versionMin)
	expectedMin := fmt.Sprintf("account/%s/events/%010d", account, versionMin)

	versionMax := math.MaxInt32
	pathMax := EventPath(account, versionMax)
	expectedMax := fmt.Sprintf("account/%s/events/%010d", account, versionMax)

	assert.Equal(t, expectedMin, pathMin)
	assert.Equal(t, expectedMax, pathMax)
}

func TestSnapshotPath(t *testing.T) {
	account := "account_3"

	versionMin := 0
	pathMin := SnapshotPath(account, versionMin)
	expectedMin := fmt.Sprintf("account/%s/snapshot/%010d", account, versionMin)

	versionMax := math.MaxInt32
	pathMax := SnapshotPath(account, versionMax)
	expectedMax := fmt.Sprintf("account/%s/snapshot/%010d", account, versionMax)

	assert.Equal(t, expectedMin, pathMin)
	assert.Equal(t, expectedMax, pathMax)
}

func BenchmarkSnapshotPath(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		SnapshotPath("X", 0)
	}
}

func BenchmarkSnapshotsPath(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		SnapshotsPath("X")
	}
}

func BenchmarkEventPath(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		EventPath("X", 0)
	}
}

func BenchmarkEventsPath(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		EventsPath("X")
	}
}
