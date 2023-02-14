package iri_test

import (
	"testing"

	"github.com/contomap/iri"
)

func TestIRI_NormalizePercentEncoding(t *testing.T) {
	tt := []struct {
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
			in:   `https://é.example.org`,
			want: `https://é.example.org`,
		},
		{
			name: "Preserve percent encoding when it is necessary",
			in:   `https://é.example.org/dog%20house/%c2%B5`,
			want: `https://é.example.org/dog%20house/µ`,
		},
		{
			name: "Example from https://github.com/google/xtoproto/issues/23",
			in:   "https://wiktionary.org/wiki/%E1%BF%AC%CF%8C%CE%B4%CE%BF%CF%82",
			want: `https://wiktionary.org/wiki/Ῥόδος`,
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			in, err := iri.Parse(tc.in)
			if err != nil {
				t.Errorf("IRI %s is not a valid IRI: %v", tc.in, err)
			}
			// TODO (type-rework) Parse() already normalizes -- extract test; also: add cases for non-path encoding
			if got := in.NormalizePercentEncoding(); got.String() != tc.want {
				t.Errorf("NormalizePercentEncoding(%q) = \n  %s, want\n  %s", tc.in, got, tc.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tt := []struct {
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
			name: "https://example.org/#André",
			in:   `https://example.org/#André`,
			want: `https://example.org/#André`,
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
			name:    "https://example.org/#André then some whitespace",
			in:      "https://example.org/#André then some whitespace",
			want:    ``,
			wantErr: true,
		},
		{
			name: "query with iprivate character",
			in:   "https://example.org?\ue000",
			want: `https://example.org?`,
		},
		{
			// This is an "intentional" parse error; It is to showcase that examples from RFC 3987 with XML notation
			// must not be taken literally. As per chapter 1.4 (page 5), these are meant to escape the "only US-ASCII"
			// RFC text. This particular example comes from page 12.
			name:    "XML notation from RFC to represent non-ascii characters",
			in:      `https://r&#xE9;sum&#xE9;.example.org`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid utf-8 B5",
			in:      `https://é.example.org/dog%20house/%B5`,
			want:    ``,
			wantErr: true,
		},
		{
			// 181 is not a valid utf-8 octal. Check out https://www.utf8-chartable.de/.
			name:    "invalid utf-8 B5",
			in:      `https://é.example.org/dog%20house/%20%b5`,
			want:    ``,
			wantErr: true,
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := iri.Parse(tc.in)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Errorf("got err %v, wantErr = %v", err, tc.wantErr)
			}
			if got.String() != tc.want {
				t.Errorf("Parse(%q) got %s, want %s", tc.in, got, tc.want)
			}
		})
	}
}

func TestResolveReference(t *testing.T) {
	tt := []struct {
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
			base: "https://red@google.com:341",
			ref:  `https://example/q?abc=1&def=2`,
			want: `https://example/q?abc=1&def=2`,
		},
		// {
		// 	name: "An empty same document reference \"\" resolves against the URI part of the base URI; any fragment part is ignored. See Uniform Resource Identifiers (URI) [RFC3986]",
		// 	base: "http://bigbird@google.com/path#x-frag",
		// 	ref:  ``,
		// 	want: `http://bigbird@google.com/path`,
		// },
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			base, err := iri.Parse(tc.base)
			if err != nil {
				t.Errorf("base IRI %s is not a valid IRI: %v", tc.base, err)
			}
			ref, err := iri.Parse(tc.ref)
			if err != nil {
				t.Errorf("ref IRI %s is not a valid IRI: %v", tc.ref, err)
			}
			got := base.ResolveReference(ref).String()
			if got != tc.want {
				t.Errorf("ResolveReference(%s, %s) got\n  %s, want\n  %s", tc.base, tc.ref, got, tc.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	tt := []struct {
		value string
	}{
		{""},
		{"example.com"},
		{"example.com:"},
		{"example.com:0"},
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
		{"https://example.com:22/path/to?"},
		{"https://example.com:22/path/to?#"},
		{"https://example.com:22/path/to?#frag"},
		{"https://example.com:22/path/to?q=a"},
		{"https://example.com:22/path/to?q=a#b"},
		{"https://example.com:22/path/to?q=a#"},
		{"https://#"},
		{`http://example/q?abc=1&def=2`},

		{"scheme:opaque?query#fragment"},
		{"scheme://userinfo@host/path?query#fragment"},

		{"https://@example.com"},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.value, func(t *testing.T) {
			t.Parallel()
			got, err := iri.Parse(tc.value)
			if err != nil {
				t.Errorf("Parse() return error: got: %v", err)
			}
			if got.String() != tc.value {
				t.Errorf("Parse().String() roundtrip failed:\n  input:  %s\n  output: %s\n  parts:\n%#v", tc.value, got, got)
			}
		})
	}
}

func TestStringFromCreatedObject(t *testing.T) {
	tt := []struct {
		in   iri.IRI
		want string
	}{
		{in: iri.IRI{}, want: ""},

		{in: iri.IRI{ForceFragment: true}, want: "#"},
		{in: iri.IRI{ForceFragment: true, Fragment: "forcedFragment"}, want: "#forcedFragment"},
		{in: iri.IRI{ForceFragment: false, Fragment: "loneFragment"}, want: "#loneFragment"},

		{in: iri.IRI{ForceQuery: true}, want: "?"},
		{in: iri.IRI{ForceQuery: true, Query: "q=forced"}, want: "?q=forced"},
		{in: iri.IRI{ForceQuery: false, Query: "q=lone"}, want: "?q=lone"},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			got := tc.in.String()
			if got != tc.want {
				t.Errorf("String() mismatch: got: '%s', want: '%s'", got, tc.want)
			}
			err := tc.in.Check()
			if err != nil {
				t.Errorf("Check() returned error: %v", err)
			}
		})
	}
}

type verifyFunc func(testing.TB, iri.IRI)

func TestProperties(t *testing.T) {
	tt := []struct {
		in     string
		verify verifyFunc
	}{
		{in: "", verify: is(iri.IRI{})},
		{in: "test:", verify: hasScheme("test")},
		{in: "test:example.com:1234", verify: hasScheme("test")},
		{in: "//example.com:1234", verify: allOf(hasScheme(""), hasHost("example.com:1234"))},
		{in: "https://user:pwd@example.com", verify: allOf(hasScheme("https"), hasUser("user:pwd"), hasHost("example.com"))},
		{in: "https://@example.com", verify: allOf(hasScheme("https"), hasUser(""), hasHost("example.com"))},

		{in: "//[2001:db8:3333:4444:5555:6666:7777:8888]", verify: hasHost("[2001:db8:3333:4444:5555:6666:7777:8888]")},
		{in: "//[2001:0db8:0001:0000:0000:0ab9:C0A8:0102]", verify: hasHost("[2001:0db8:0001:0000:0000:0ab9:C0A8:0102]")},
		{in: "//[2001:db8::]", verify: hasHost("[2001:db8::]")},
		{in: "//[::1234:5678]", verify: hasHost("[::1234:5678]")},
		{in: "//[2001:db8::1234:5678]", verify: hasHost("[2001:db8::1234:5678]")},
		{in: "//[2001:0db8:0001:0000:0000:0ab9:C0A8:0102]", verify: hasHost("[2001:0db8:0001:0000:0000:0ab9:C0A8:0102]")},
		{in: "//[2001:db8:1::ab9:C0A8:102]", verify: hasHost("[2001:db8:1::ab9:C0A8:102]")},

		{in: "//[2001:db8:3333:4444:5555:6666:1.2.3.4]", verify: hasHost("[2001:db8:3333:4444:5555:6666:1.2.3.4]")},
		{in: "//[::11.22.33.44]", verify: hasHost("[::11.22.33.44]")},
		{in: "//[2001:db8::123.123.123.123]", verify: hasHost("[2001:db8::123.123.123.123]")},
		{in: "//[::1234:5678:91.123.4.56]", verify: hasHost("[::1234:5678:91.123.4.56]")},
		{in: "//[::1234:5678:1.2.3.4]", verify: hasHost("[::1234:5678:1.2.3.4]")},
		{in: "//[2001:db8::1234:5678:5.6.7.8]", verify: hasHost("[2001:db8::1234:5678:5.6.7.8]")},

		{in: "//[::2:3:4:5:6:7:8]", verify: hasHost("[::2:3:4:5:6:7:8]")},
		{in: "//[::3:4:5:6:7:8]", verify: hasHost("[::3:4:5:6:7:8]")},
		{in: "//[::4:5:6:7:8]", verify: hasHost("[::4:5:6:7:8]")},
		{in: "//[::5:6:7:8]", verify: hasHost("[::5:6:7:8]")},
		{in: "//[::6:7:8]", verify: hasHost("[::6:7:8]")},
		{in: "//[::7:8]", verify: hasHost("[::7:8]")},
		{in: "//[::8]", verify: hasHost("[::8]")},
		{in: "//[::]", verify: hasHost("[::]")},
		{in: "//[7::]", verify: hasHost("[7::]")},
		{in: "//[6:7::]", verify: hasHost("[6:7::]")},
		{in: "//[5:6:7::]", verify: hasHost("[5:6:7::]")},
		{in: "//[4:5:6:7::]", verify: hasHost("[4:5:6:7::]")},
		{in: "//[3:4:5:6:7::]", verify: hasHost("[3:4:5:6:7::]")},
		{in: "//[2:3:4:5:6:7::]", verify: hasHost("[2:3:4:5:6:7::]")},
		{in: "//[2:3:4:5:6:7::]", verify: hasHost("[2:3:4:5:6:7::]")},
		{in: "//[1:2:3:4:5:6:7::]", verify: hasHost("[1:2:3:4:5:6:7::]")},
		{in: "//[1:2:3:4:5:6::8]", verify: hasHost("[1:2:3:4:5:6::8]")},
		{in: "//[1:2:3:4:5::7:8]", verify: hasHost("[1:2:3:4:5::7:8]")},
		{in: "//[1:2:3:4::6:7:8]", verify: hasHost("[1:2:3:4::6:7:8]")},
		{in: "//[1:2:3::5:6:7:8]", verify: hasHost("[1:2:3::5:6:7:8]")},
		{in: "//[1:2::4:5:6:7:8]", verify: hasHost("[1:2::4:5:6:7:8]")},
		{in: "//[1::3:4:5:6:7:8]", verify: hasHost("[1::3:4:5:6:7:8]")},
		{in: "//[1::4:5:6:7:8]", verify: hasHost("[1::4:5:6:7:8]")},
		{in: "//[1::5:6:7:8]", verify: hasHost("[1::5:6:7:8]")},
		{in: "//[1::6:7:8]", verify: hasHost("[1::6:7:8]")},
		{in: "//[1::7:8]", verify: hasHost("[1::7:8]")},
		{in: "//[1::8]", verify: hasHost("[1::8]")},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			got, err := iri.Parse(tc.in)
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}
			tc.verify(t, got)
			if t.Failed() {
				t.Logf("got: %#v", got)
			}
		})
	}
}

func hasScheme(expected string) verifyFunc {
	return func(t testing.TB, got iri.IRI) {
		if got.Scheme != expected {
			t.Errorf("invalid scheme. want: '%v'", expected)
		}
	}
}

func hasHost(expected string) verifyFunc {
	return func(t testing.TB, got iri.IRI) {
		if got.Host != expected {
			t.Errorf("invalid host. want: '%v'", expected)
		}
	}
}

func hasUser(expected string) verifyFunc {
	return func(t testing.TB, got iri.IRI) {
		if got.UserInfo != expected {
			t.Errorf("invalid user info. want: '%v'", expected)
		}
	}
}

func is(expected iri.IRI) verifyFunc {
	return func(t testing.TB, got iri.IRI) {
		if got != expected {
			t.Errorf("is not matching. want: %#v", expected)
		}
	}
}

func allOf(list ...verifyFunc) verifyFunc {
	return func(t testing.TB, got iri.IRI) {
		for _, entry := range list {
			entry(t, got)
		}
	}
}
