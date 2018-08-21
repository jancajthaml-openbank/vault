package utils

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

var pathsTestParams = RunParams{
	Setup: SetupParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	},
	Journal: JournalParams{},
	Metrics: MetricsParams{},
}

var pathsBenchmarkParams = RunParams{
	Setup: SetupParams{
		Tenant:      "tenant",
		RootStorage: "/benchmark",
	},
	Journal: JournalParams{},
	Metrics: MetricsParams{},
}

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
	pathMin := EventPath(pathsTestParams, account, versionMin)
	expectedMin := fmt.Sprintf("%s/account/%s/events/%010d", pathsTestParams.Setup.RootStorage, account, versionMin)

	versionMax := math.MaxInt32
	pathMax := EventPath(pathsTestParams, account, versionMax)
	expectedMax := fmt.Sprintf("%s/account/%s/events/%010d", pathsTestParams.Setup.RootStorage, account, versionMax)

	assert.Equal(t, expectedMin, pathMin)
	assert.Equal(t, expectedMax, pathMax)
}

func TestSnapshotPath(t *testing.T) {
	account := "account_3"

	versionMin := 0
	pathMin := SnapshotPath(pathsTestParams, account, versionMin)
	expectedMin := fmt.Sprintf("%s/account/%s/snapshot/%010d", pathsTestParams.Setup.RootStorage, account, versionMin)

	versionMax := math.MaxInt32
	pathMax := SnapshotPath(pathsTestParams, account, versionMax)
	expectedMax := fmt.Sprintf("%s/account/%s/snapshot/%010d", pathsTestParams.Setup.RootStorage, account, versionMax)

	assert.Equal(t, expectedMin, pathMin)
	assert.Equal(t, expectedMax, pathMax)
}

func TestMetadataPath(t *testing.T) {
	account := "account_4"

	actual := MetadataPath(pathsTestParams, account)
	expected := fmt.Sprintf("%s/account/%s/meta", pathsTestParams.Setup.RootStorage, account)

	assert.Equal(t, expected, actual)
}

func BenchmarkSnapshotPath(b *testing.B) {

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		SnapshotPath(pathsBenchmarkParams, "X", 0)
	}
}

func BenchmarkSnapshotsPath(b *testing.B) {

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		SnapshotsPath(pathsBenchmarkParams, "X")
	}
}

func BenchmarkEventPath(b *testing.B) {

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		EventPath(pathsBenchmarkParams, "X", 0)
	}
}

func BenchmarkEventsPath(b *testing.B) {

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		EventsPath(pathsBenchmarkParams, "X")
	}
}

func BenchmarkMetadataPath(b *testing.B) {

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		MetadataPath(pathsBenchmarkParams, "X")
	}
}
