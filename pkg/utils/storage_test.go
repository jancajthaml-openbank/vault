package utils

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPad(version int) string {
	return fmt.Sprintf("%010d", version)
}

func TestListDirectory(t *testing.T) {

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

	require.Nil(t, os.MkdirAll(tmpdir, os.ModePerm))
	defer os.RemoveAll(tmpdir)

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

func TestCountFiles(t *testing.T) {
	tmpdir := "/tmp/test_storage"

	require.Nil(t, os.MkdirAll(tmpdir, os.ModePerm))
	defer os.RemoveAll(tmpdir)

	for i := 0; i < 60; i++ {
		file, err := os.Create(tmpdir + "/" + testPad(i) + "F")
		require.Nil(t, err)
		file.Close()
	}

	for i := 0; i < 40; i++ {
		err := os.MkdirAll(tmpdir+"/"+testPad(i)+"D", os.ModePerm)
		require.Nil(t, err)
	}

	numberOfFiles := CountFiles(tmpdir)
	assert.Equal(t, 60, numberOfFiles)
}

func BenchmarkCountFiles(b *testing.B) {
	tmpdir := "/tmp/test_storage/"

	os.MkdirAll(tmpdir, os.ModePerm)
	defer os.RemoveAll(tmpdir)

	for i := 0; i < 1000; i++ {
		file, err := os.Create(fmt.Sprintf("%s%010d", tmpdir, i))
		require.Nil(b, err)
		file.Close()
	}

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		CountFiles(tmpdir)
	}
}

func BenchmarkListDirectory(b *testing.B) {
	tmpdir := "/tmp/test_storage/"

	os.MkdirAll(tmpdir, os.ModePerm)
	defer os.RemoveAll(tmpdir)

	for i := 0; i < 1000; i++ {
		file, err := os.Create(fmt.Sprintf("%s%010d", tmpdir, i))
		require.Nil(b, err)
		file.Close()
	}

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		ListDirectory(tmpdir, true)
	}
}

func BenchmarkExists(b *testing.B) {
	tmpdir := "/tmp/test_storage/"
	filename := tmpdir + "exist"

	os.MkdirAll(tmpdir, os.ModePerm)
	defer os.RemoveAll(tmpdir)
	file, err := os.Create(filename)
	require.Nil(b, err)
	file.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		require.True(b, Exists(filename))
	}
}

func BenchmarkUpdateFile(b *testing.B) {
	tmpdir := "/tmp/test_storage/"
	filename := tmpdir + "updated"

	os.MkdirAll(tmpdir, os.ModePerm)
	defer os.RemoveAll(tmpdir)
	file, err := os.Create(filename)
	require.Nil(b, err)
	file.Close()

	data := []byte("abcd")

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		require.True(b, UpdateFile(filename, data))
	}
}

func BenchmarkAppendFile(b *testing.B) {
	tmpdir := "/tmp/test_storage/"
	filename := tmpdir + "appended"

	os.MkdirAll(tmpdir, os.ModePerm)
	defer os.RemoveAll(tmpdir)
	file, err := os.Create(filename)
	require.Nil(b, err)
	file.Close()

	data := []byte("abcd")

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		require.True(b, AppendFile(filename, data))
	}
}

func BenchmarkReadFileFully(b *testing.B) {
	tmpdir := "/tmp/test_storage/"
	filename := tmpdir + "appended"

	os.MkdirAll(tmpdir, os.ModePerm)
	defer os.RemoveAll(tmpdir)
	file, err := os.Create(filename)
	require.Nil(b, err)
	file.Close()

	require.True(b, UpdateFile(filename, []byte("abcd")))

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		ok, _ := ReadFileFully(filename)
		require.True(b, ok)
	}
}
