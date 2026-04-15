// Package filter provides log line filtering based on include/exclude patterns.
package filter

import (
	"regexp"
)

// Filter holds compiled include and exclude patterns.
type Filter struct {
	include []*regexp.Regexp
	exclude []*regexp.Regexp
}

// Config holds raw pattern strings for building a Filter.
type Config struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

// New compiles the patterns in cfg and returns a Filter.
// Returns an error if any pattern fails to compile.
func New(cfg Config) (*Filter, error) {
	inc, err := compileAll(cfg.Include)
	if err != nil {
		return nil, err
	}
	exc, err := compileAll(cfg.Exclude)
	if err != nil {
		return nil, err
	}
	return &Filter{include: inc, exclude: exc}, nil
}

// Allow reports whether line passes the filter.
// A line is allowed when:
//   - it matches at least one include pattern (or no include patterns are set), AND
//   - it does not match any exclude pattern.
func (f *Filter) Allow(line string) bool {
	if len(f.exclude) > 0 {
		for _, re := range f.exclude {
			if re.MatchString(line) {
				return false
			}
		}
	}
	if len(f.include) == 0 {
		return true
	}
	for _, re := range f.include {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

func compileAll(patterns []string) ([]*regexp.Regexp, error) {
	out := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		out = append(out, re)
	}
	return out, nil
}
