// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"io"
	"os"
	"path/filepath"
	"sort"
)

// ListDirectory returns sorted slice of item names in given absolute path
// default sorting is ascending
func ListDirectory(absPath string, ascending bool) []string {
	f, err := os.Open(absPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	list, err := f.Readdir(-1)
	if err != nil {
		return nil
	}

	v := make([]string, len(list))

	for i, value := range list {
		v[i] = filepath.Base(value.Name())
	}

	if ascending {
		sort.Slice(v, func(i, j int) bool {
			return v[i] < v[j]
		})
	} else {
		sort.Slice(v, func(i, j int) bool {
			return v[i] > v[j]
		})
	}

	return v
}

// CountNodes returns number of items in directory
func CountNodes(absPath string) int {
	f, err := os.Open(absPath)
	if err != nil {
		return -1
	}
	defer f.Close()

	list, err := f.Readdir(-1)
	if err != nil {
		return -1
	}

	return len(list)
}

// Exists returns true if absolute path exists
func Exists(absPath string) bool {
	_, err := os.Stat(absPath)
	return err == nil
}

// TouchFile creates files given absolute path if file does not already exist
func TouchFile(absPath string) bool {
	if err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm); err != nil {
		return false
	}

	f, err := os.OpenFile(absPath, os.O_RDONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if err != nil {
		return false
	}
	defer f.Close()

	return true
}

// ReadFile reads size bytes of file given absolute path and size
func ReadFile(absPath string, size int) (bool, []byte) {
	f, err := os.OpenFile(absPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return false, nil
	}
	defer f.Close()

	buf := make([]byte, size)
	_, err = f.Read(buf)
	if err != nil && err != io.EOF {
		return false, nil
	}

	return true, buf
}

// ReadFileFully reads whole file given absolute path
func ReadFileFully(absPath string) (bool, []byte) {
	f, err := os.OpenFile(absPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return false, nil
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false, nil
	}

	buf := make([]byte, fi.Size())
	_, err = f.Read(buf)
	if err != nil && err != io.EOF {
		return false, nil
	}

	return true, buf
}

// WriteFile writes data given absolute path to a file if that file does not
// already exists
func WriteFile(absPath string, data []byte) bool {
	if err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm); err != nil {
		return false
	}

	f, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if err != nil {
		return false
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return false
	}

	return true
}

// UpdateFile rewrite file with data given absolute path to a file if that file
// exist
func UpdateFile(absPath string, data []byte) bool {
	f, err := os.OpenFile(absPath, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return false
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return false
	}

	return true
}

// AppendFile appens data given absolute path to a file, creates it if it does
// not exist
func AppendFile(absPath string, data []byte) bool {
	if err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm); err != nil {
		return false
	}

	f, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return false
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return false
	}

	return true
}
