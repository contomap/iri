package iri_test

import (
	"testing"

	"github.com/contomap/iri"
)

func TestIRI_NormalizePercentEncoding(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "a",
			in:   `https://github.com/google/xtoproto/testing#prop1`,
			want: "https://github.com/google/xtoproto/testing#prop1",
		},
		{
			name: "b",
			in:   `https://github.com/google/xtoproto/testing#prop1`,
			want: "https://github.com/google/xtoproto/testing#prop1",
		},
		{
			name: "non ascii é",
			in:   `http://é.example.org`,
			want: `http://é.example.org`,
		},
		{
			name: "Preserve percent encoding when it is necessary",
			in:   `http://é.example.org/dog%20house/%c2%B5`,
			want: `http://é.example.org/dog%20house/µ`,
		},
		{
			name: "Example from https://github.com/google/xtoproto/issues/23",
			in:   "http://wiktionary.org/wiki/%E1%BF%AC%CF%8C%CE%B4%CE%BF%CF%82",
			want: `http://wiktionary.org/wiki/Ῥόδος`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in, err := iri.Parse(tt.in)
			if err != nil {
				t.Errorf("IRI %s is not a valid IRI: %v", tt.in, err)
			}
			if got := in.NormalizePercentEncoding(); got.String() != tt.want {
				t.Errorf("NormalizePercentEncoding(%q) = \n  %s, want\n  %s", tt.in, got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "prop1",
			in:   `https://github.com/google/xtoproto/testing#prop1`,
			want: "https://github.com/google/xtoproto/testing#prop1",
		},
		{
			name: "http://example.org/#André",
			in:   `http://example.org/#André`,
			want: `http://example.org/#André`,
		},
		{
			name: "valid urn:uuid",
			in:   `urn:uuid:6c689097-8097-4421-9def-05e835f2dbb8`,
			want: `urn:uuid:6c689097-8097-4421-9def-05e835f2dbb8`,
		},
		{
			name: "valid urn:uuid:",
			in:   `urn:uuid:`,
			want: `urn:uuid:`,
		},
		{
			name: "valid a:b:c:",
			in:   `a:b:c:`,
			want: `a:b:c:`,
		},
		{
			name:    "http://example.org/#André then some whitespace",
			in:      "http://example.org/#André then some whitespace",
			want:    ``,
			wantErr: true,
		},
		{
			// This is an "intentional" parse error; It is to showcase that examples from RFC 3987 with XML notation
			// must not be taken literally. As per chapter 1.4 (page 5), these are meant to escape the "only US-ASCII"
			// RFC text. This particular example comes from page 12.
			name:    "XML notation from RFC to represent non-ascii characters",
			in:      `http://r&#xE9;sum&#xE9;.example.org`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid utf-8 B5",
			in:      `http://é.example.org/dog%20house/%B5`,
			want:    ``,
			wantErr: true,
		},
		{
			// 181 is not a valid utf-8 octal. Check out https://www.utf8-chartable.de/.
			name:    "invalid utf-8 B5",
			in:      `http://é.example.org/dog%20house/%20%b5`,
			want:    ``,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := iri.Parse(tt.in)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf("got err %v, wantErr = %v", err, tt.wantErr)
			}
			if got.String() != tt.want {
				t.Errorf("Parse(%q) got %s, want %s", tt.in, got, tt.want)
			}
		})
	}
}

func TestResolveReference(t *testing.T) {
	tests := []struct {
		name      string
		base, ref string
		want      string
	}{
		{
			name: "prop1",
			base: `https://github.com/google/xtoproto/testing#prop1`,
			ref:  `#3`,
			want: "https://github.com/google/xtoproto/testing#3",
		},
		{
			name: "slash blah",
			base: `https://github.com/google/xtoproto/testing#prop1`,
			ref:  `/blah`,
			want: "https://github.com/blah",
		},
		{
			name: "empty ref",
			base: `https://github.com/google/xtoproto/testing#prop1`,
			ref:  ``,
			want: "https://github.com/google/xtoproto/testing#prop1",
		},
		{
			name: "different full iri",
			base: `https://github.com/google/xtoproto/testing#prop1`,
			ref:  `http://x`,
			want: "http://x",
		},
		{
			name: "blank fragment",
			base: `https://github.com/google/xtoproto/testing`,
			ref:  `#`,
			want: "https://github.com/google/xtoproto/testing#",
		},
		{
			name: "replace completely",
			base: "http://red@google.com:341",
			ref:  `http://example/q?abc=1&def=2`,
			want: `http://example/q?abc=1&def=2`,
		},
		// {
		// 	name: "An empty same document reference \"\" resolves against the URI part of the base URI; any fragment part is ignored. See Uniform Resource Identifiers (URI) [RFC3986]",
		// 	base: "http://bigbird@google.com/path#x-frag",
		// 	ref:  ``,
		// 	want: `http://bigbird@google.com/path`,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, err := iri.Parse(tt.base)
			if err != nil {
				t.Errorf("base IRI %s is not a valid IRI: %v", tt.base, err)
			}
			ref, err := iri.Parse(tt.ref)
			if err != nil {
				t.Errorf("ref IRI %s is not a valid IRI: %v", tt.ref, err)
			}
			got := base.ResolveReference(ref).String()
			if got != tt.want {
				t.Errorf("ResolveReference(%s, %s) got\n  %s, want\n  %s", tt.base, tt.ref, got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		value string
	}{
		{""},
		{"example.com"},
		{"example.com:22"},
		{"example.com:22/path/to"},
		{"example.com:22/path/to?"},
		{"example.com:22/path/to?q=a"},
		{"example.com:22/path/to?q=a#b"},
		{"example.com:22/path/to?q=a#"},
		{"#"},
		{""},
		{"https://example.com"},
		{"https://example.com:22"},
		{"https://example.com:22/path/to"},
		{"https://example.com:22/path/to?q=a"},
		{"https://example.com:22/path/to?q=a#b"},
		{"https://example.com:22/path/to?q=a#"},
		{"https://#"},
		{`http://example/q?abc=1&def=2`},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			iri, err := iri.Parse(tt.value)
			if err != nil {
				t.Errorf("Parse() return error: got: %v", err)
			}
			if iri.String() != tt.value {
				t.Errorf(".parts().toIRI() roundtrip failed:\n  input:  %s\n  output: %s\n  parts:\n%#v", tt.value, iri, iri)
			}
		})
	}
}
