package iri //nolint: testpackage

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestMustCompileNamedPanics(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			got := fmt.Sprintf("%v", p)
			if !strings.HasPrefix(got, "failed to compile regexp example:") {
				t.Errorf("expected specific panic text. got: %q", got)
			}
		} else {
			t.Errorf("expected panic")
		}
	}()
	mustCompileNamed("example", "[")
}

func TestRegExps(t *testing.T) {
	tests := []struct {
		name string
		re   *regexp.Regexp
		in   string
		want bool
	}{
		{
			name: "space is not a valid iri character",
			re:   iunreservedRE,
			in:   " ",
			want: false,
		},
		{
			name: "þ is unreserved",
			re:   iunreservedRE,
			in:   "\u00FE",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.re.MatchString(tt.in)
			if got != tt.want {
				t.Errorf("%s.Match(%q) got %v, want %v", tt.re, tt.in, got, tt.want)
			}
		})
	}
}
