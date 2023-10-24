// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package touchbundle

import (
	"reflect"
	"strings"
	"unicode"
)

const snakeCaseSeparator rune = '_'

type snakifier struct {
	output       strings.Builder
	parsingUpper bool
	token        []rune
}

func (s *snakifier) push(r rune) {
	switch {
	case r == snakeCaseSeparator:
		// separators in the identifiers are preserved if they
		// aren't leading or trailing
		s.flush()

	case len(s.token) == 0:
		s.parsingUpper = unicode.IsUpper(r) || unicode.IsTitle(r)
		s.token = append(s.token, unicode.ToLower(r))

	case unicode.IsLower(r):
		if s.parsingUpper {
			s.parsingUpper = false
			if len(s.token) > 1 {
				// this ends a run of capitals
				// consider the last capital as the start of another token
				last := s.token[len(s.token)-1]
				s.token = s.token[0 : len(s.token)-1]
				s.flush()
				s.token = append(s.token, last)
			}

			// if len(token) == 1, it was the start of a token that we are now parsing
		}

		s.token = append(s.token, r)

	case unicode.IsUpper(r) || unicode.IsTitle(r):
		if !s.parsingUpper {
			// this is the start of a new token
			s.flush()
			s.parsingUpper = true
		}

		s.token = append(s.token, unicode.ToLower(r))

	default:
		// non-letters, non-separators
		s.token = append(s.token, r)
	}
}

func (s *snakifier) flush() {
	if len(s.token) > 0 {
		if s.output.Len() > 0 {
			s.output.WriteRune(snakeCaseSeparator)
		}

		s.output.WriteString(string(s.token))
		s.token = s.token[:0]
	}
}

func (s *snakifier) String() string {
	return s.output.String()
}

// toSnakeCase converts a golang identifier, e.g. a struct field name,
// into snake case.
func toSnakeCase(identifier string) string {
	if len(identifier) == 0 {
		return identifier
	}

	s := snakifier{
		token: make([]rune, 0, 15),
	}

	for _, r := range identifier {
		s.push(r)
	}

	s.flush()
	return s.String()
}

// MetricName determines the metric name of a struct field.  A metric name is generated
// by first looking at the TagName struct field tag, failling back to the snakecase
// of the field name if that tag is not provide or is empty.  The first occurrence of "*"
// in the tag will be replaced by the snakecase of the field name, allowing for easy
// prefixes and suffixes.
//
// For example:
//
//	type Bundle struct {
//	    // metric name is:  something_count
//	    SomethingCount *prometheus.CounterVec `labelNames:"foo,bar"`
//
//	    // metric name is: prefix_my_gauge
//	    MyGauge *prometheus.GaugeVec `name:"prefix_*" labelNames:"foo,bar"`
//
//	    // metric name is: custom_name
//	    AnotherCounter prometheus.Counter `name:"custom_name"`
//	}
func MetricName(f reflect.StructField) string {
	snakeCase := toSnakeCase(f.Name)
	name := strings.Replace(
		f.Tag.Get(TagName),
		"*",
		snakeCase,
		1,
	)

	if len(name) == 0 {
		name = snakeCase
	}

	return name
}
