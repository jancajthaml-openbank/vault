package utils

import (
	"fmt"
	"math"
	"strconv"
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
	params := RunParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	}
	account := "account_2"

	versionMin := 0
	pathMin := EventPath(params, account, versionMin)
	expectedMin := fmt.Sprintf("%s/account/%s/events/%010d", params.RootStorage, account, versionMin)

	versionMax := math.MaxInt32
	pathMax := EventPath(params, account, versionMax)
	expectedMax := fmt.Sprintf("%s/account/%s/events/%010d", params.RootStorage, account, versionMax)

	assert.Equal(t, expectedMin, pathMin)
	assert.Equal(t, expectedMax, pathMax)
}

func TestSnapshotPath(t *testing.T) {
	params := RunParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	}
	account := "account_3"

	versionMin := 0
	pathMin := SnapshotPath(params, account, versionMin)
	expectedMin := fmt.Sprintf("%s/account/%s/snapshot/%010d", params.RootStorage, account, versionMin)

	versionMax := math.MaxInt32
	pathMax := SnapshotPath(params, account, versionMax)
	expectedMax := fmt.Sprintf("%s/account/%s/snapshot/%010d", params.RootStorage, account, versionMax)

	assert.Equal(t, expectedMin, pathMin)
	assert.Equal(t, expectedMax, pathMax)
}

func TestMetadataPath(t *testing.T) {
	params := RunParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	}
	account := "account_4"

	actual := MetadataPath(params, account)
	expected := fmt.Sprintf("%s/account/%s/meta", params.RootStorage, account)

	assert.Equal(t, expected, actual)
}

func BenchmarkItoa(b *testing.B) {
	testItoa := func(input int) string {
		return strconv.Itoa(input)
	}

	for n := 0; n < b.N; n++ {
		testItoa(1)
	}
}

func BenchmarkFmtPad(b *testing.B) {
	testPad := func(input int) string {
		return fmt.Sprintf("%010d", input)
	}

	for n := 0; n < b.N; n++ {
		testPad(1)
	}
}

func BenchmarkSnapshotPath(b *testing.B) {
	params := RunParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	}

	for n := 0; n < b.N; n++ {
		SnapshotPath(params, "X", 0)
	}
}

func BenchmarkSnapshotsPath(b *testing.B) {
	params := RunParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	}

	for n := 0; n < b.N; n++ {
		SnapshotsPath(params, "X")
	}
}

func BenchmarkEventPath(b *testing.B) {
	params := RunParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	}

	for n := 0; n < b.N; n++ {
		EventPath(params, "X", 0)
	}
}

func BenchmarkEventsPath(b *testing.B) {
	params := RunParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	}

	for n := 0; n < b.N; n++ {
		EventsPath(params, "X")
	}
}

func BenchmarkMetadataPath(b *testing.B) {
	params := RunParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	}

	for n := 0; n < b.N; n++ {
		MetadataPath(params, "X")
	}
}
