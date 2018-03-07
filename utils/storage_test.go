package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"io/ioutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	removeContents := func(dir string) {
		d, err := os.Open(dir)
		if err != nil {
			return
		}
		defer d.Close()
		names, err := d.Readdirnames(-1)
		if err != nil {
			return
		}
		for _, name := range names {
			err = os.RemoveAll(filepath.Join(dir, name))
			if err != nil {
				return
			}
		}
		return
	}

	tmpdir := "/tmp/bench_storage"

	removeContents(tmpdir)

	for i := 0; i < 100; i++ {
		var file, _ = os.Create(fmt.Sprintf("%s/file_%d", tmpdir, i))
		file.Close()
	}
}

func TestListDirectory(t *testing.T) {
	testPad := func(version int) string {
		return fmt.Sprintf("%010d", version)
	}

	tmpdir := "/tmp/test_storage"

	NewSlice := func(start, end, step int) []int {
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

func BenchmarkListDirectory(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ListDirectory("/tmp/bench_storage", true)
	}
}

func BenchmarkUpdateFile(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		UpdateFile("/tmp/bench_storage/file_1", []byte("data"))
	}
}

func BenchmarkReadFileFully(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ReadFileFully("/tmp/bench_storage/file_1")
	}
}

func BenchmarkIOUtilReadFile(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ioutil.ReadFile("/tmp/bench_storage/file_1")
	}
}
