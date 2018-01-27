package utils

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewSlice(start, end, step int) []int {
	if step <= 0 || end < start {
		return []int{}
	}
	s := make([]int, 0, 1+(end-start)/step)
	for start <= end {
		s = append(s, start)
		start += step
	}
	return s
}

func TestListDirectory(t *testing.T) {
	testPad := func(version int) string {
		return fmt.Sprintf("%010d", version)
	}

	tmpdir := "/tmp/test_storage"

	// FIXME check for error
	os.MkdirAll(tmpdir, os.ModePerm)

	items := NewSlice(0, 10, 1)

	for _, i := range items {
		var file, _ = os.Create(tmpdir + "/" + testPad(i))
		file.Close()
	}

	list := ListDirectory(tmpdir, true)

	require.NotNil(t, list)
	assert.Equal(t, len(items), len(list))
	assert.Equal(t, testPad(items[0]), list[0])
	assert.Equal(t, testPad(items[len(items)-1]), list[len(list)-1])
}
