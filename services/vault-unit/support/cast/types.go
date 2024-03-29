// Copyright (c) 2016-2023, Jan Cajthaml <jan.cajthaml@gmail.com>
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

package cast

import (
	"fmt"
	"reflect"
	"unsafe"
)

// BytesToString casts []byte type to string, this does not copy the original
// value so if original slice is changed string will also change
func BytesToString(bytes []byte) string {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: sliceHeader.Data,
		Len:  sliceHeader.Len,
	}))
}

// StringToBytes casts string type to []byte, this does not copy the original
// value so if original string is changed []byte will also change
func StringToBytes(s string) []byte {
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: stringHeader.Data,
		Len:  stringHeader.Len,
		Cap:  stringHeader.Len,
	}))
}

// StringToPositiveInteger converts natural numeric string to int64
func StringToPositiveInteger(s string) (int64, error) {
	l := len(s)
	if l == 0 {
		return 0, fmt.Errorf("not a number")
	}
	x := int64(0)
	for i := 0; i < l; i++ {
		switch s[i] {
		case '0':
			x = x * 10
		case '1':
			x = x*10 + 1
		case '2':
			x = x*10 + 2
		case '3':
			x = x*10 + 3
		case '4':
			x = x*10 + 4
		case '5':
			x = x*10 + 5
		case '6':
			x = x*10 + 6
		case '7':
			x = x*10 + 7
		case '8':
			x = x*10 + 8
		case '9':
			x = x*10 + 9
		default:
			return 0, fmt.Errorf("invalid digit %q", s[i])
		}
	}
	return x, nil
}
