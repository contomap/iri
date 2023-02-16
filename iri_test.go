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
			in:   "https://example.com/sub/path/testing#frag1",
			want: "https://example.com/sub/path/testing#frag1",
		},
		{
			name: "https://example.org/#André",
			in:   "https://example.org/#André",
			want: "https://example.org/#André",
		},
		{
			name: "valid urn:uuid",
			in:   "urn:uuid:6c689097-8097-4421-9def-05e835f2dbb8",
			want: "urn:uuid:6c689097-8097-4421-9def-05e835f2dbb8",
		},
		{
			name: "valid urn:uuid:",
			in:   "urn:uuid:",
			want: "urn:uuid:",
		},
		{
			name: "valid a:b:c:",
			in:   "a:b:c:",
			want: "a:b:c:",
		},
		{
			name:    "https://example.org/#André then some whitespace",
			in:      "https://example.org/#André then some whitespace",
			want:    "",
			wantErr: true,
		},
		{
			name: "query with iprivate character",
			in:   "https://example.org?\ue000",
			want: "https://example.org?",
		},
		{
			// This is an "intentional" parse error; It is to showcase that examples from RFC 3987 with XML notation
			// must not be taken literally. As per chapter 1.4 (page 5), these are meant to escape the "only US-ASCII"
			// RFC text. This particular example comes from page 12.
			name:    "XML notation from RFC to represent non-ascii characters",
			in:      "https://r&#xE9;sum&#xE9;.example.org",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid utf-8 B5",
			in:      "https://é.example.org/dog%20house/%B5",
			want:    "",
			wantErr: true,
		},
		{
			// 181 is not a valid utf-8 octal. Check out https://www.utf8-chartable.de/.
			name:    "invalid utf-8 B5",
			in:      "https://é.example.org/dog%20house/%20%b5",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid scheme",
			in:      " :",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid authority",
			in:      "//[not-a-v6]",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid path",
			in:      "/ ",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid query",
			in:      "? ",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid fragment",
			in:      "# ",
			want:    "",
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
				t.Errorf("Parse(%q) got %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseRFC3986Samples(t *testing.T) {
	tt := []struct {
		value string
	}{
		{"ftp://ftp.is.co.za/rfc/rfc1808.txt"},
		{"https://www.ietf.org/rfc/rfc2396.txt"},
		{"ldap://[2001:db8::7]/c=GB?objectClass?one"},
		{"mailto:John.Doe@example.com"},
		{"news:comp.infosystems.www.servers.unix"},
		{"tel:+1-816-555-1212"},
		{"telnet://192.0.2.16:80/"},
		{"urn:oasis:names:specification:docbook:dtd:xml:4.1.2"},
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
				t.Errorf("Parse().String() roundtrip failed:\n  input:  %q\n  output: %q\n  parts:\n%#v", tc.value, got, got)
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
		{"http://example/q?abc=1&def=2"},

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
				t.Errorf("Parse().String() roundtrip failed:\n  input:  %q\n  output: %q\n  parts:\n%#v", tc.value, got, got)
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
				t.Errorf("String() mismatch: got: %q, want: %q", got, tc.want)
			}
			_, err := iri.Parse(got)
			if err != nil {
				t.Errorf("Parse(got) returned error: got: %q, err: %v", got, err)
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
			in:   "https://example.com/sub/path/testing#frag1",
			want: "https://example.com/sub/path/testing#frag1",
		},
		{
			name: "b",
			in:   "https://example.com/sub/path/testing#frag1",
			want: "https://example.com/sub/path/testing#frag1",
		},
		{
			name: "non ascii é",
			in:   "https://é.example.org",
			want: "https://é.example.org",
		},
		{
			name: "encoded userinfo",
			in:   "https://%c2%B5@example.org",
			want: "https://µ@example.org",
		},
		{
			name: "encoded host",
			in:   "https://%c2%B5.example.org",
			want: "https://µ.example.org",
		},
		{
			name: "Preserve percent encoding when it is necessary",
			in:   "https://é.example.org/dog%20house/%c2%B5",
			want: "https://é.example.org/dog%20house/µ",
		},
		{
			name: "encoded query",
			in:   "https://example.org?q=%c2%B5",
			want: "https://example.org?q=µ",
		},
		{
			name: "encoded fragment",
			in:   "https://example.org#%c2%B5",
			want: "https://example.org#µ",
		},
		{
			name: "Example from https://github.com/google/xtoproto/issues/23",
			in:   "https://wiktionary.org/wiki/%E1%BF%AC%CF%8C%CE%B4%CE%BF%CF%82",
			want: "https://wiktionary.org/wiki/Ῥόδος",
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			in, err := iri.Parse(tc.in)
			if err != nil {
				t.Errorf("IRI %q is not a valid IRI: %v", tc.in, err)
			}
			if got, _ := iri.NormalizePercentEncoding(in); got.String() != tc.want {
				t.Errorf("NormalizePercentEncoding(%q) = \n  %q, want\n  %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizePercentEncodingErrors(t *testing.T) {
	tt := []struct {
		value iri.IRI
	}{
		{value: iri.IRI{Authority: "%FF"}},
		{value: iri.IRI{Path: "%FF"}},
		{value: iri.IRI{Query: "%FF"}},
		{value: iri.IRI{Fragment: "%FF"}},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.value.String(), func(t *testing.T) {
			t.Parallel()
			if _, err := iri.NormalizePercentEncoding(tc.value); err == nil {
				t.Errorf("Parse() did not return an error")
			}
		})
	}
}

func TestResolveReferenceManualSamples(t *testing.T) {
	// Many test cases here duplicate the samples from the RFCs, yet they are kept here as a manual test basis.
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
		{
			name: "relative unavailable path",
			base: "https://example.com",
			ref:  "../../nowhere",
			want: "https://example.com/nowhere",
		},
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
			name: "empty ref as per RFC 3986 algorithm in 5.2.2.",
			base: "https://example.com/sub/path/testing#frag1",
			ref:  "",
			want: "https://example.com/sub/path/testing",
		},
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
				t.Errorf("base IRI %q is not a valid IRI: %v", tc.base, err)
			}
			ref, err := iri.Parse(tc.ref)
			if err != nil {
				t.Errorf("ref IRI %q is not a valid IRI: %v", tc.ref, err)
			}
			got := base.ResolveReference(ref).String()
			if got != tc.want {
				t.Errorf("ResolveReference(%q, %q)\n  got %q\n want %q", tc.base, tc.ref, got, tc.want)
			}
		})
	}
}

func TestResolveReferenceRFC3986Samples(t *testing.T) {
	base, baseErr := iri.Parse("http://a/b/c/d;p?q")
	if baseErr != nil {
		t.Fatalf("Base IRI is not correct: %v", baseErr)
	}
	tt := []struct {
		ref  string
		want string
	}{
		// Normal examples as per 5.4.1
		{ref: "g:h", want: "g:h"},
		{ref: "g", want: "http://a/b/c/g"},
		{ref: "./g", want: "http://a/b/c/g"},
		{ref: "g/", want: "http://a/b/c/g/"},
		{ref: "/g", want: "http://a/g"},
		{ref: "//g", want: "http://g"},
		{ref: "?y", want: "http://a/b/c/d;p?y"},
		{ref: "g?y", want: "http://a/b/c/g?y"},
		{ref: "#s", want: "http://a/b/c/d;p?q#s"},
		{ref: "g#s", want: "http://a/b/c/g#s"},
		{ref: "g?y#s", want: "http://a/b/c/g?y#s"},
		{ref: ";x", want: "http://a/b/c/;x"},
		{ref: "g;x", want: "http://a/b/c/g;x"},
		{ref: "g;x?y#s", want: "http://a/b/c/g;x?y#s"},
		{ref: "", want: "http://a/b/c/d;p?q"},
		{ref: ".", want: "http://a/b/c/"},
		{ref: "./", want: "http://a/b/c/"},
		{ref: "..", want: "http://a/b/"},
		{ref: "../", want: "http://a/b/"},
		{ref: "../g", want: "http://a/b/g"},
		{ref: "../..", want: "http://a/"},
		{ref: "../../", want: "http://a/"},
		{ref: "../../g", want: "http://a/g"},

		// Abnormal examples as per 5.4.2
		{ref: "../../../g", want: "http://a/g"},
		{ref: "../../../../g", want: "http://a/g"},
		{ref: "/./g", want: "http://a/g"},
		{ref: "/../g", want: "http://a/g"},
		{ref: "g.", want: "http://a/b/c/g."},
		{ref: ".g", want: "http://a/b/c/.g"},
		{ref: "g..", want: "http://a/b/c/g.."},
		{ref: "..g", want: "http://a/b/c/..g"},

		{ref: "./../g", want: "http://a/b/g"},
		{ref: "./g/.", want: "http://a/b/c/g/"},
		{ref: "g/./h", want: "http://a/b/c/g/h"},
		{ref: "g/../h", want: "http://a/b/c/h"},
		{ref: "g;x=1/./y", want: "http://a/b/c/g;x=1/y"},
		{ref: "g;x=1/../y", want: "http://a/b/c/y"},

		{ref: "g?y/./x", want: "http://a/b/c/g?y/./x"},
		{ref: "g?y/../x", want: "http://a/b/c/g?y/../x"},
		{ref: "g#s/./x", want: "http://a/b/c/g#s/./x"},
		{ref: "g#s/../x", want: "http://a/b/c/g#s/../x"},

		{ref: "http:g", want: "http:g"},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.ref, func(t *testing.T) {
			t.Parallel()
			ref, err := iri.Parse(tc.ref)
			if err != nil {
				t.Errorf("ref IRI %q is not a valid IRI: %v", tc.ref, err)
			}
			got := base.ResolveReference(ref).String()
			if got != tc.want {
				t.Errorf("ResolveReference(%q, %q)\n  got %q\n want %q", base.String(), tc.ref, got, tc.want)
			}
		})
	}
}

func TestResolveReferenceRFC1808Samples(t *testing.T) {
	// Although many samples of RFC 1808 are similar to that of RFC 3986,
	// those of RFC 1808 do use a base that contains a fragment.
	// However, RFC 3986 also obsoletes RFC 1808, so some samples are not valid.
	base, baseErr := iri.Parse("http://a/b/c/d;p?q#f")
	if baseErr != nil {
		t.Fatalf("Base IRI is not correct: %v", baseErr)
	}
	tt := []struct {
		ref  string
		want string
	}{
		// Normal examples as per 5.1
		{ref: "g:h", want: "g:h"},
		{ref: "g", want: "http://a/b/c/g"},
		{ref: "./g", want: "http://a/b/c/g"},
		{ref: "g/", want: "http://a/b/c/g/"},
		{ref: "/g", want: "http://a/g"},
		{ref: "//g", want: "http://g"},
		{ref: "?y", want: "http://a/b/c/d;p?y"},
		{ref: "g?y", want: "http://a/b/c/g?y"},
		{ref: "g?y/./x", want: "http://a/b/c/g?y/./x"},
		{ref: "#s", want: "http://a/b/c/d;p?q#s"},
		{ref: "g#s", want: "http://a/b/c/g#s"},
		{ref: "g#s/./x", want: "http://a/b/c/g#s/./x"},
		{ref: "g?y#s", want: "http://a/b/c/g?y#s"},
		{ref: ";x", want: "http://a/b/c/;x"},
		{ref: "g;x", want: "http://a/b/c/g;x"},
		{ref: "g;x?y#s", want: "http://a/b/c/g;x?y#s"},
		{ref: ".", want: "http://a/b/c/"},
		{ref: "./", want: "http://a/b/c/"},
		{ref: "..", want: "http://a/b/"},
		{ref: "../", want: "http://a/b/"},
		{ref: "../g", want: "http://a/b/g"},
		{ref: "../..", want: "http://a/"},
		{ref: "../../", want: "http://a/"},
		{ref: "../../g", want: "http://a/g"},

		// Abnormal examples as per 5.2
		// {ref: "", want: "http://a/b/c/d;p?q#f"}, // superseded by RFC 3986, algorithm in 5.2.2.
		{ref: "", want: "http://a/b/c/d;p?q"}, // the current behaviour as per RFC 3986, algorithm in 5.2.2.

		// {ref: "../../../g", want: "http://a/../g"}, // superseded by RFC 3986 as per examples
		// {ref: "../../../../g", want: "http://a/../../g"}, // superseded by RFC 3986 as per examples
		// {ref: "/./g", want: "http://a/./g"}, // superseded by RFC 3986 as per examples
		// {ref: "/../g", want: "http://a/../g"}, // superseded by RFC 3986 as per examples
		{ref: "g.", want: "http://a/b/c/g."},
		{ref: ".g", want: "http://a/b/c/.g"},
		{ref: "g..", want: "http://a/b/c/g.."},
		{ref: "..g", want: "http://a/b/c/..g"},

		{ref: "./../g", want: "http://a/b/g"},
		{ref: "./g/.", want: "http://a/b/c/g/"},
		{ref: "g/./h", want: "http://a/b/c/g/h"},
		{ref: "g/../h", want: "http://a/b/c/h"},

		{ref: "http:g", want: "http:g"},
		{ref: "http:", want: "http:"},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.ref, func(t *testing.T) {
			t.Parallel()
			ref, err := iri.Parse(tc.ref)
			if err != nil {
				t.Errorf("ref IRI %q is not a valid IRI: %v", tc.ref, err)
			}
			got := base.ResolveReference(ref).String()
			if got != tc.want {
				t.Errorf("ResolveReference(%q, %q)\n  got %q\n want %q", base.String(), tc.ref, got, tc.want)
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

			t.Errorf("invalid scheme. want: %q", expected)
		}
	}
}

func hasAuthority(expected string) verifyFunc {
	return func(t testing.TB, got iri.IRI) {
		if got.Authority != expected {
			t.Errorf("invalid host. want: %q", expected)
		}
	}
}

func hasPath(expected string) verifyFunc {
	return func(t testing.TB, got iri.IRI) {
		if got.Path != expected {
			t.Errorf("invalid path. want: %q", expected)
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
