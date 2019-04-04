// Copyright © 2019 Moises P. Sena <moisespsena@gmail.com>
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

package cmd

import (
	"path/filepath"
	"strings"

	"github.com/moisespsena-go/xbindata"
)

// parseRecursive determines whether the given path has a recursive indicator and
// returns a new path with the recursive indicator chopped off if it does.
//
//  ex:
//      /path/to/foo/...    -> (/path/to/foo, true)
//      /path/to/bar        -> (/path/to/bar, false)
func parseInput(path string) xbindata.InputConfig {
	if strings.HasSuffix(path, "/...") {
		return xbindata.InputConfig{
			Path:      filepath.Clean(path[:len(path)-4]),
			Recursive: true,
		}
	} else {
		return xbindata.InputConfig{
			Path:      filepath.Clean(path),
			Recursive: false,
		}
	}

}

// StringsValue implements the flag.Value interface and allows multiple
// calls to the same variable to append a list.
type StringsValue struct {
	m      map[string]interface{}
	values []string
	setter func(i int, value string)
	i      int
}

func (s *StringsValue) Type() string {
	return "strings"
}

func (s *StringsValue) String() string {
	return strings.Join(s.values, ", ")
}

func (s *StringsValue) Set(value string) error {
	if s.m == nil {
		s.m = map[string]interface{}{}
	}
	if _, ok := s.m[value]; !ok {
		s.m[value] = nil
		s.values = append(s.values, value)
		s.setter(s.i, value)
	}
	s.i++
	return nil
}
