package ignore

import (
	"fmt"
	"regexp"

	"github.com/gobwas/glob"
)

type IgnoreSlice []string

func (s IgnoreSlice) Items() (r []*regexp.Regexp, err error) {
	for j, pattern := range s {
		if regexPattern, err := regexp.Compile(pattern); err != nil {
			return nil, fmt.Errorf("invalid regex pattern #%d (%q): %v", j, pattern, err)
		} else {
			r = append(r, regexPattern)
		}
	}
	return
}

type IgnoreGlobSlice []string

func (s IgnoreGlobSlice) Items() (r []glob.Glob, err error) {
	for j, pattern := range s {
		if globPattern, err := glob.Compile(pattern); err != nil {
			return nil, fmt.Errorf("invalid glob pattern #%d (%q): %v", j, pattern, err)
		} else {
			r = append(r, globPattern)
		}
	}
	return
}
