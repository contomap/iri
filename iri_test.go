package iri_test

import (
	"testing"

	"github.com/contomap/iri"
)

func TestParse(t *testing.T) {
	tt := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "prop1",
			in:   `https://example.com/sub/path/testing#frag1`,
			want: "https://example.com/sub/path/testing#frag1",
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
			_, err := iri.Parse(got)
			if err != nil {
				t.Errorf("Parse(got) returned error: got: '%s', err: %v", got, err)
			}
		})
	}
}

func TestNormalizePercentEncoding(t *testing.T) {
	tt := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "a",
			in:   `https://example.com/sub/path/testing#frag1`,
			want: "https://example.com/sub/path/testing#frag1",
		},
		{
			name: "b",
			in:   `https://example.com/sub/path/testing#frag1`,
			want: "https://example.com/sub/path/testing#frag1",
		},
		{
			name: "non ascii é",
			in:   `https://é.example.org`,
			want: `https://é.example.org`,
		},
		{
			name: "encoded userinfo",
			in:   `https://%c2%B5@example.org`,
			want: `https://µ@example.org`,
		},
		{
			name: "encoded host",
			in:   `https://%c2%B5.example.org`,
			want: `https://µ.example.org`,
		},
		{
			name: "Preserve percent encoding when it is necessary",
			in:   `https://é.example.org/dog%20house/%c2%B5`,
			want: `https://é.example.org/dog%20house/µ`,
		},
		{
			name: "encoded query",
			in:   `https://example.org?q=%c2%B5`,
			want: `https://example.org?q=µ`,
		},
		{
			name: "encoded fragment",
			in:   `https://example.org#%c2%B5`,
			want: `https://example.org#µ`,
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
			if got, _ := iri.NormalizePercentEncoding(in); got.String() != tc.want {
				t.Errorf("NormalizePercentEncoding(%q) = \n  %s, want\n  %s", tc.in, got, tc.want)
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
			name: "different scheme",
			base: "https://example.com/sub/path/example?q=123#frag1",
			ref:  "email:user@example.com",
			want: "email:user@example.com",
		},
		{
			name: "different authority",
			base: "https://user@example.com?q=123#frag1",
			ref:  "//example/q?abc=1&def=2",
			want: "https://example/q?abc=1&def=2",
		},
		{
			name: "absolute path",
			base: "https://example.com/sub/path/testing#frag1",
			ref:  "/abs",
			want: "https://example.com/abs",
		},
		{
			name: "absolute path with details",
			base: "https://example.com/sub/path/testing#frag1",
			ref:  "/abs?q=123#frag2",
			want: "https://example.com/abs?q=123#frag2",
		},
		{
			name: "relative parent path",
			base: "https://example.com/sub/path/testing#frag1",
			ref:  "../other",
			want: "https://example.com/sub/other",
		},
		{
			name: "relative local relative path",
			base: "https://example.com/sub/path/testing#frag1",
			ref:  "./here",
			want: "https://example.com/sub/path/here",
		},
		{
			name: "relative local path",
			base: "https://example.com/sub/path/testing#frag1",
			ref:  ".",
			want: "https://example.com/sub/path/",
		},
		/* TODO (type-rework) - this probably needs to have the "beginning slash" re-introduced, conditionally
		{
			name: "relative unavailable path",
			base: "https://example.com",
			ref:  "../../nowhere",
			want: "https://example.com/nowhere",
		},
		*/
		{
			name: "different query",
			base: "https://example.com/sub/path/testing#frag1",
			ref:  "?q2=abc",
			want: "https://example.com/sub/path/testing?q2=abc",
		},
		{
			name: "different fragment",
			base: "https://example.com/sub/path/testing#frag1",
			ref:  "#frag3",
			want: "https://example.com/sub/path/testing#frag3",
		},
		{
			name: "blank fragment",
			base: "https://example.com/sub/path/testing",
			ref:  "#",
			want: "https://example.com/sub/path/testing#",
		},
		{
			name: "empty ref",
			base: "https://example.com/sub/path/testing#frag1",
			ref:  "",
			want: "https://example.com/sub/path/testing#frag1",
		},
		// {
		// 	name: "An empty same document reference \"\" resolves against the URI part of the base URI; any fragment part is ignored. See Uniform Resource Identifiers (URI) [RFC3986]",
		// 	base: "http://user@example.com/path#x-frag",
		// 	ref:  "",
		// 	want: `http://user@example.com/path`,
		// },
		{
			name: "empty to empty",
			base: "",
			ref:  "",
			want: "",
		},
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

func TestProperties(t *testing.T) {
	tt := []struct {
		in     string
		verify verifyFunc
	}{
		{in: "", verify: is(iri.IRI{})},
		{in: "test:", verify: hasScheme("test")},
		{in: "test:example.com:1234", verify: hasScheme("test")},
		{in: "//example.com:1234", verify: allOf(hasScheme(""), hasAuthority("example.com:1234"))},
		{in: "https://user:pwd@example.com", verify: allOf(hasScheme("https"), hasAuthority("user:pwd@example.com"))},
		{in: "https://@example.com", verify: allOf(hasScheme("https"), hasAuthority("@example.com"))},

		{in: "//[2001:db8:3333:4444:5555:6666:7777:8888]", verify: hasAuthority("[2001:db8:3333:4444:5555:6666:7777:8888]")},
		{in: "//[2001:0db8:0001:0000:0000:0ab9:C0A8:0102]", verify: hasAuthority("[2001:0db8:0001:0000:0000:0ab9:C0A8:0102]")},
		{in: "//[2001:db8::]", verify: hasAuthority("[2001:db8::]")},
		{in: "//[::1234:5678]", verify: hasAuthority("[::1234:5678]")},
		{in: "//[2001:db8::1234:5678]", verify: hasAuthority("[2001:db8::1234:5678]")},
		{in: "//[2001:0db8:0001:0000:0000:0ab9:C0A8:0102]", verify: hasAuthority("[2001:0db8:0001:0000:0000:0ab9:C0A8:0102]")},
		{in: "//[2001:db8:1::ab9:C0A8:102]", verify: hasAuthority("[2001:db8:1::ab9:C0A8:102]")},

		{in: "//[2001:db8:3333:4444:5555:6666:1.2.3.4]", verify: hasAuthority("[2001:db8:3333:4444:5555:6666:1.2.3.4]")},
		{in: "//[::11.22.33.44]", verify: hasAuthority("[::11.22.33.44]")},
		{in: "//[2001:db8::123.123.123.123]", verify: hasAuthority("[2001:db8::123.123.123.123]")},
		{in: "//[::1234:5678:91.123.4.56]", verify: hasAuthority("[::1234:5678:91.123.4.56]")},
		{in: "//[::1234:5678:1.2.3.4]", verify: hasAuthority("[::1234:5678:1.2.3.4]")},
		{in: "//[2001:db8::1234:5678:5.6.7.8]", verify: hasAuthority("[2001:db8::1234:5678:5.6.7.8]")},

		{in: "//[::2:3:4:5:6:7:8]", verify: hasAuthority("[::2:3:4:5:6:7:8]")},
		{in: "//[::3:4:5:6:7:8]", verify: hasAuthority("[::3:4:5:6:7:8]")},
		{in: "//[::4:5:6:7:8]", verify: hasAuthority("[::4:5:6:7:8]")},
		{in: "//[::5:6:7:8]", verify: hasAuthority("[::5:6:7:8]")},
		{in: "//[::6:7:8]", verify: hasAuthority("[::6:7:8]")},
		{in: "//[::7:8]", verify: hasAuthority("[::7:8]")},
		{in: "//[::8]", verify: hasAuthority("[::8]")},
		{in: "//[::]", verify: hasAuthority("[::]")},
		{in: "//[7::]", verify: hasAuthority("[7::]")},
		{in: "//[6:7::]", verify: hasAuthority("[6:7::]")},
		{in: "//[5:6:7::]", verify: hasAuthority("[5:6:7::]")},
		{in: "//[4:5:6:7::]", verify: hasAuthority("[4:5:6:7::]")},
		{in: "//[3:4:5:6:7::]", verify: hasAuthority("[3:4:5:6:7::]")},
		{in: "//[2:3:4:5:6:7::]", verify: hasAuthority("[2:3:4:5:6:7::]")},
		{in: "//[2:3:4:5:6:7::]", verify: hasAuthority("[2:3:4:5:6:7::]")},
		{in: "//[1:2:3:4:5:6:7::]", verify: hasAuthority("[1:2:3:4:5:6:7::]")},
		{in: "//[1:2:3:4:5:6::8]", verify: hasAuthority("[1:2:3:4:5:6::8]")},
		{in: "//[1:2:3:4:5::7:8]", verify: hasAuthority("[1:2:3:4:5::7:8]")},
		{in: "//[1:2:3:4::6:7:8]", verify: hasAuthority("[1:2:3:4::6:7:8]")},
		{in: "//[1:2:3::5:6:7:8]", verify: hasAuthority("[1:2:3::5:6:7:8]")},
		{in: "//[1:2::4:5:6:7:8]", verify: hasAuthority("[1:2::4:5:6:7:8]")},
		{in: "//[1::3:4:5:6:7:8]", verify: hasAuthority("[1::3:4:5:6:7:8]")},
		{in: "//[1::4:5:6:7:8]", verify: hasAuthority("[1::4:5:6:7:8]")},
		{in: "//[1::5:6:7:8]", verify: hasAuthority("[1::5:6:7:8]")},
		{in: "//[1::6:7:8]", verify: hasAuthority("[1::6:7:8]")},
		{in: "//[1::7:8]", verify: hasAuthority("[1::7:8]")},
		{in: "//[1::8]", verify: hasAuthority("[1::8]")},

		{in: "tel:7042;phone-context=example.com", verify: allOf(hasScheme("tel"), hasPath("7042;phone-context=example.com"))},
		{in: "email:user@example.com", verify: allOf(hasScheme("email"), hasPath("user@example.com"))},
		{in: "/", verify: hasPath("/")},
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

type verifyFunc func(testing.TB, iri.IRI)

func hasScheme(expected string) verifyFunc {
	return func(t testing.TB, got iri.IRI) {
		if got.Scheme != expected {
			t.Errorf("invalid scheme. want: '%v'", expected)
		}
	}
}

func hasAuthority(expected string) verifyFunc {
	return func(t testing.TB, got iri.IRI) {
		if got.Authority != expected {
			t.Errorf("invalid host. want: '%v'", expected)
		}
	}
}

func hasPath(expected string) verifyFunc {
	return func(t testing.TB, got iri.IRI) {
		if got.Path != expected {
			t.Errorf("invalid path. want: '%v'", expected)
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
